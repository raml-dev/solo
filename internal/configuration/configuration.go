package configuration

type Configuration struct {
	Collection CollectionConfiguration
}

type CollectionConfiguration struct {
	Path string
}
