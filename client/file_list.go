package client

// FileList represents a file_list.yml file downloaded from server
type FileList struct {
	Version        string      `yaml:"version"`
	DownloadPrefix string      `yaml:"downloadprefix"`
	Deletes        []FileEntry `yaml:"deletes"`
	Downloads      []FileEntry `yaml:"downloads"`
	Unpacks        []FileEntry `yaml:"unpacks"`
}

// FileEntry is an entry inside FileList
type FileEntry struct {
	Name string `yaml:"name"`
	Md5  string `yaml:"md5"`
	Date string `yaml:"date"`
	Zip  string `yaml:"zip"`
	Size int    `yaml:"size"`
}
