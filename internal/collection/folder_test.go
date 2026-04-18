package collection

import "testing"

func TestFolderGetRequestByIdBasic(t *testing.T) {
	req := Request{Id: "req-1", Name: "test request"}
	folder := Folder{
		Requests: []Request{req},
	}

	got, err := folder.GetRequestById("req-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got == nil {
		t.Fatal("expected request, got nil")
	}

	if got.Id != "req-1" {
		t.Fatalf("expected req-1, got %s", got.Id)
	}
}

func TestFolderGetRequestByIdFromSecondSubfolder(t *testing.T) {
	root := Folder{
		Id: "root",
		Folders: []Folder{
			{
				Id:       "sub-1",
				Requests: []Request{{Id: "req-1", Name: "first"}},
			},
			{
				Id:       "sub-2",
				Requests: []Request{{Id: "req-2", Name: "second"}},
			},
		},
	}

	got, err := root.GetRequestById("req-2")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got == nil {
		t.Fatal("expected request, got nil")
	}

	if got.Id != "req-2" {
		t.Fatalf("expected req-2, got %s", got.Id)
	}
}

func TestAddRequestToFolderSubfolder(t *testing.T) {
	root := Folder{
		Id: "root",
		Folders: []Folder{
			{Id: "sub-1"},
		},
	}

	err := root.AddRequestToFolder("sub-1", Request{Id: "req-sub", Name: "request in subfolder"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(root.Folders[0].Requests) != 1 {
		t.Fatalf("expected 1 request in subfolder, got %d", len(root.Folders[0].Requests))
	}

	if root.Folders[0].Requests[0].Id != "req-sub" {
		t.Fatalf("expected req-sub, got %s", root.Folders[0].Requests[0].Id)
	}
}

func TestFolderGetRequestByIdThreeLevelNesting(t *testing.T) {
	root := Folder{
		Id: "root",
		Folders: []Folder{
			{
				Id: "level-1",
				Folders: []Folder{
					{
						Id: "level-2",
						Folders: []Folder{
							{
								Id:       "level-3",
								Requests: []Request{{Id: "req-l3", Name: "third-level"}},
							},
						},
					},
				},
			},
		},
	}

	got, err := root.GetRequestById("req-l3")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got == nil {
		t.Fatal("expected request, got nil")
	}

	if got.Id != "req-l3" {
		t.Fatalf("expected req-l3, got %s", got.Id)
	}
}

func TestRemoveRequestFromFolderSubfolder(t *testing.T) {
	root := Folder{
		Id: "root",
		Folders: []Folder{
			{
				Id:       "sub-1",
				Requests: []Request{{Id: "req-a", Name: "to-remove"}, {Id: "req-b", Name: "to-keep"}},
			},
		},
	}

	err := root.RemoveRequestFromFolder("sub-1", "req-a")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(root.Folders[0].Requests) != 1 {
		t.Fatalf("expected 1 request in subfolder, got %d", len(root.Folders[0].Requests))
	}

	if root.Folders[0].Requests[0].Id != "req-b" {
		t.Fatalf("expected remaining request req-b, got %s", root.Folders[0].Requests[0].Id)
	}
}

func TestUpdateRequestInFolderSubfolder(t *testing.T) {
	root := Folder{
		Id: "root",
		Folders: []Folder{
			{
				Id: "sub-1",
				Requests: []Request{
					{
						Id:      "req-upd",
						Name:    "old-name",
						Verb:    "GET",
						Url:     "https://example.com/old",
						Body:    "{}",
						Headers: map[string]string{"x-old": "1"},
					},
				},
			},
		},
	}

	err := root.UpdateRequestInFolder("sub-1", Request{
		Id:      "req-upd",
		Name:    "new-name",
		Verb:    "POST",
		Url:     "https://example.com/new",
		Body:    "{\"ok\":true}",
		Headers: map[string]string{"x-new": "1"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated := root.Folders[0].Requests[0]
	if updated.Name != "new-name" {
		t.Fatalf("expected name new-name, got %s", updated.Name)
	}
	if updated.Verb != "POST" {
		t.Fatalf("expected verb POST, got %s", updated.Verb)
	}
	if updated.Url != "https://example.com/new" {
		t.Fatalf("expected new url, got %s", updated.Url)
	}
	if updated.Headers["x-new"] != "1" {
		t.Fatalf("expected header x-new=1, got %v", updated.Headers)
	}
}
