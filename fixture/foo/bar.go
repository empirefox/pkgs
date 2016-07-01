package foo

import "github.com/jinzhu/gorm"

type Bar struct {
	gorm.Model
	Name string `multi 
				 line
				 tag`
}

type IBar interface {
	GetName() string
}

type Sbar *Bar
