package entity

import "errors"

var (
	ErrNoPermission   = errors.New("dont have permissions")
	ErrFolderExist    = errors.New("folder already exist")
	ErrComicsExist    = errors.New("comics already exist")
	ErrComicsNotExist = errors.New("comics doesnt exist")
	ErrFolderNotExist = errors.New("folder doesnt exist")
)
