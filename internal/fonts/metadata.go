// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package fonts

import (
	"encoding/binary"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"golang.org/x/text/encoding/charmap"
)

var errNoUsableFontMetadata = errors.New("fonts: no usable font metadata")

const (
	maxFontTables = 256

	sfntTrueTypeVersion = 0x00010000
	sfntOpenTypeCFF     = 0x4f54544f // "OTTO"
	sfntAppleTrueType   = 0x74727565 // "true"
	sfntCollection      = 0x74746366 // "ttcf"

	sfntNameTable = 0x6e616d65 // "name"
	sfntPostTable = 0x706f7374 // "post"

	sfntNameIDFamily             = 1
	sfntNameIDTypographicFamily  = 16
	sfntPlatformIDMacintosh      = 1
	sfntPlatformIDWindows        = 3
	sfntEncodingIDMacintoshRoman = 0
	sfntEncodingIDWindowsUCS2    = 1
)

type sfntTableRecord struct {
	offset int64
	length uint32
}

func parseFontFile(path string) ([]SystemFont, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".ttc" {
		return parseFontCollectionMetadata(file)
	}

	parsed, err := parseFontMetadataAt(file, 0)
	if err != nil {
		return nil, err
	}
	return []SystemFont{parsed}, nil
}

func parseFontCollectionMetadata(file io.ReaderAt) ([]SystemFont, error) {
	header, err := readFontBytes(file, 0, 12)
	if err != nil {
		return nil, err
	}
	if binary.BigEndian.Uint32(header) != sfntCollection {
		return nil, errNoUsableFontMetadata
	}

	numFonts := binary.BigEndian.Uint32(header[8:])
	if numFonts == 0 || numFonts > maxFontTables {
		return nil, errNoUsableFontMetadata
	}

	offsetData, err := readFontBytes(file, 12, int(numFonts)*4)
	if err != nil {
		return nil, err
	}

	fonts := make([]SystemFont, 0, numFonts)
	for i := 0; i < int(numFonts); i++ {
		offset := int64(binary.BigEndian.Uint32(offsetData[i*4:]))
		font, err := parseFontMetadataAt(file, offset)
		if err != nil {
			continue
		}
		fonts = append(fonts, font)
	}
	if len(fonts) == 0 {
		return nil, errNoUsableFontMetadata
	}
	return fonts, nil
}

func parseFontMetadataAt(file io.ReaderAt, offset int64) (SystemFont, error) {
	tables, err := parseFontTableDirectory(file, offset)
	if err != nil {
		return SystemFont{}, err
	}

	nameTable, ok := tables[sfntNameTable]
	if !ok {
		return SystemFont{}, errNoUsableFontMetadata
	}

	family, err := fontFamilyNameFromTable(file, nameTable)
	if err != nil {
		return SystemFont{}, err
	}

	isMonospace := false
	if postTable, ok := tables[sfntPostTable]; ok {
		isMonospace, err = fontIsFixedPitch(file, postTable)
		if err != nil {
			return SystemFont{}, err
		}
	}

	return SystemFont{
		Family:      family,
		IsMonospace: isMonospace,
	}, nil
}

func parseFontTableDirectory(file io.ReaderAt, offset int64) (map[uint32]sfntTableRecord, error) {
	header, err := readFontBytes(file, offset, 12)
	if err != nil {
		return nil, err
	}

	switch binary.BigEndian.Uint32(header) {
	case sfntTrueTypeVersion, sfntOpenTypeCFF, sfntAppleTrueType:
	default:
		return nil, errNoUsableFontMetadata
	}

	numTables := int(binary.BigEndian.Uint16(header[4:]))
	if numTables == 0 || numTables > maxFontTables {
		return nil, errNoUsableFontMetadata
	}

	records, err := readFontBytes(file, offset+12, numTables*16)
	if err != nil {
		return nil, err
	}

	tables := make(map[uint32]sfntTableRecord, numTables)
	for i := 0; i < numTables; i++ {
		record := records[i*16:]
		tag := binary.BigEndian.Uint32(record)
		tables[tag] = sfntTableRecord{
			offset: int64(binary.BigEndian.Uint32(record[8:])),
			length: binary.BigEndian.Uint32(record[12:]),
		}
	}
	return tables, nil
}

func fontFamilyNameFromTable(file io.ReaderAt, table sfntTableRecord) (string, error) {
	for _, id := range []uint16{sfntNameIDTypographicFamily, sfntNameIDFamily} {
		name, err := fontName(file, table, id)
		if err != nil {
			continue
		}
		name = strings.TrimSpace(name)
		if name != "" {
			return name, nil
		}
	}

	return "", errNoUsableFontMetadata
}

func fontName(file io.ReaderAt, table sfntTableRecord, id uint16) (string, error) {
	const headerSize, entrySize = 6, 12
	if table.length < headerSize {
		return "", errNoUsableFontMetadata
	}

	header, err := readFontBytes(file, table.offset, headerSize)
	if err != nil {
		return "", err
	}
	count := int(binary.BigEndian.Uint16(header[2:]))
	stringOffset := int(binary.BigEndian.Uint16(header[4:]))
	if table.length < headerSize+uint32(count*entrySize) {
		return "", errNoUsableFontMetadata
	}

	records, err := readFontBytes(file, table.offset+headerSize, count*entrySize)
	if err != nil {
		return "", err
	}

	for i := 0; i < count; i++ {
		record := records[i*entrySize:]
		if binary.BigEndian.Uint16(record[6:]) != id {
			continue
		}

		decode, ok := nameDecoder(record)
		if !ok {
			continue
		}

		length := int(binary.BigEndian.Uint16(record[8:]))
		offset := int(binary.BigEndian.Uint16(record[10:]))
		stringStart := stringOffset + offset
		if stringStart < 0 || length < 0 || uint32(stringStart+length) > table.length {
			return "", errNoUsableFontMetadata
		}

		data, err := readFontBytes(file, table.offset+int64(stringStart), length)
		if err != nil {
			return "", err
		}
		return decode(data)
	}

	return "", errNoUsableFontMetadata
}

func nameDecoder(record []byte) (func([]byte) (string, error), bool) {
	platformID := binary.BigEndian.Uint16(record)
	encodingID := binary.BigEndian.Uint16(record[2:])

	switch {
	case platformID == sfntPlatformIDMacintosh && encodingID == sfntEncodingIDMacintoshRoman:
		return decodeMacintoshName, true
	case platformID == sfntPlatformIDWindows && encodingID == sfntEncodingIDWindowsUCS2:
		return decodeUCS2Name, true
	default:
		return nil, false
	}
}

func decodeMacintoshName(data []byte) (string, error) {
	for _, char := range data {
		if char >= 0x80 {
			decoded, err := charmap.Macintosh.NewDecoder().Bytes(data)
			if err != nil {
				return "", err
			}
			return string(decoded), nil
		}
	}
	return string(data), nil
}

func decodeUCS2Name(data []byte) (string, error) {
	if len(data)&1 != 0 {
		return "", errNoUsableFontMetadata
	}

	runes := make([]uint16, len(data)/2)
	for i := range runes {
		runes[i] = binary.BigEndian.Uint16(data[i*2:])
	}
	return string(utf16.Decode(runes)), nil
}

func fontIsFixedPitch(file io.ReaderAt, table sfntTableRecord) (bool, error) {
	const postHeaderSize = 16
	if table.length < postHeaderSize {
		return false, errNoUsableFontMetadata
	}

	header, err := readFontBytes(file, table.offset, postHeaderSize)
	if err != nil {
		return false, err
	}
	return binary.BigEndian.Uint32(header[12:]) != 0, nil
}

func readFontBytes(file io.ReaderAt, offset int64, length int) ([]byte, error) {
	if offset < 0 || length < 0 {
		return nil, errNoUsableFontMetadata
	}

	data := make([]byte, length)
	n, err := file.ReadAt(data, offset)
	if n != length {
		if err != nil {
			return nil, err
		}
		return nil, io.ErrUnexpectedEOF
	}
	return data, nil
}
