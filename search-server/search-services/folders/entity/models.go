package entity

type AddComicsIn struct {
	UserID   int64
	FolderId int64
	ComicsID int64
}

type AddComicInfo struct {
	FolderID int64
	ComicsID int64
	URL      string
}

type DeleteComicsIn struct {
	UserID   int64
	FolderId int64
	ComicsID int64
}

type Comics struct {
	ComicsID int64
	URL      string
}

type Folder struct {
	FolderID int64
	Name     string
}
