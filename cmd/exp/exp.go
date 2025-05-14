package main

import (
	"fmt"

	"github.com/TanmayKhot/pixvault/models"
)

func main() {
	gs := models.GalleryService{}
	fmt.Println(gs.Images(1))
}
