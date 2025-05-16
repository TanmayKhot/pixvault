package models

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Gallery struct {
	ID        int
	UserID    int
	Title     string
	UserEmail string
	IsPrivate bool
}

type GalleryService struct {
	DB *sql.DB
	// ImagesDir is used to tell the GalleryService where to store and locate
	// images. If not set, the GalleryService will default to using the "images"
	// directory.
	ImagesDir string
}

type Image struct {
	Path      string
	GalleryID int
	Filename  string
}

func (service *GalleryService) Create(title string, userID int) (*Gallery, error) {
	gallery := Gallery{
		Title:  title,
		UserID: userID,
	}
	row := service.DB.QueryRow(
		`INSERT INTO galleries (title, user_id)
		VALUES ($1, $2) RETURNING id;`, gallery.Title, gallery.UserID)
	err := row.Scan(&gallery.ID)
	if err != nil {
		return nil, fmt.Errorf("create gallery %w", err)
	}
	return &gallery, nil
}

func (service *GalleryService) ByID(id int) (*Gallery, error) {
	gallery := Gallery{
		ID: id,
	}
	row := service.DB.QueryRow(`
		SELECT g.title, g.user_id, u.email, g.isprivate
		FROM galleries g
		LEFT JOIN users u
		ON u.id = g.user_id
		WHERE g.id = $1;`, gallery.ID)
	err := row.Scan(&gallery.Title, &gallery.UserID, &gallery.UserEmail, &gallery.IsPrivate)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("query gallery by id :%w", err)
	}

	return &gallery, nil
}

func (service *GalleryService) ByUserID(userID int) ([]Gallery, error) {
	rows, err := service.DB.Query(`
		SELECT id, title, isprivate
		FROM galleries
		WHERE user_id = $1;`, userID)

	if err != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}

	var galleries []Gallery

	for rows.Next() {
		gallery := Gallery{
			UserID: userID,
		}
		err := rows.Scan(&gallery.ID, &gallery.Title, &gallery.IsPrivate)
		if err != nil {
			return nil, fmt.Errorf("query galleries by user: %w", err)
		}

		galleries = append(galleries, gallery)
	}

	if rows.Err() != nil {
		return nil, fmt.Errorf("query galleries by user: %w", err)
	}

	return galleries, nil
}

func (service *GalleryService) Update(gallery *Gallery) error {
	_, err := service.DB.Exec(
		`UPDATE galleries
		SET title = $2
		WHERE id = $1;`, gallery.ID, gallery.Title)

	if err != nil {
		return fmt.Errorf("update gallery: %w", err)
	}
	return nil
}

func (service *GalleryService) ChangeAccess(id int) error {
	_, err := service.DB.Exec(`
		UPDATE galleries
		SET isprivate = NOT isprivate
		WHERE id = $1;`, id)

	if err != nil {
		return fmt.Errorf("change access: %w", err)
	}
	return nil
}

func (service *GalleryService) Delete(id int) error {
	_, err := service.DB.Exec(`
		DELETE FROM galleries
		WHERE id = $1;`, id)
	if err != nil {
		return fmt.Errorf("delete gallery by id: %w", err)
	}
	return nil
}

func (service *GalleryService) galleryDir(id int) string {
	imagesDir := service.ImagesDir
	if imagesDir == "" {
		imagesDir = "images"
	}
	return filepath.Join(imagesDir, fmt.Sprintf("gallery-%d", id))
}

// List of valid extensions
func (service *GalleryService) extensions() []string {
	return []string{".jpg", ".png", ".jpeg", ".gif"}
}

// Check for valid file extension
func hasExtension(file string, extensions []string) bool {
	for _, ext := range extensions {
		file = strings.ToLower(file)
		ext = strings.ToLower(ext)
		if filepath.Ext(file) == ext {
			return true
		}
	}
	return false
}

// Fetch all valid images from the respective gallery directory
func (service *GalleryService) Images(galleryID int) ([]Image, error) {
	globPattern := filepath.Join(service.galleryDir(galleryID), "*")
	allFiles, err := filepath.Glob(globPattern)
	if err != nil {
		return nil, fmt.Errorf("retrieving gallery images: %w", err)
	}

	var images []Image
	for _, file := range allFiles {
		if hasExtension(file, service.extensions()) {
			images = append(images, Image{
				Path:      file,
				GalleryID: galleryID,
				Filename:  filepath.Base(file),
			})
		}
	}
	return images, nil
}

// Fetch a single image
func (service *GalleryService) Image(galleryID int, filename string) (Image, error) {
	imagePath := filepath.Join(service.galleryDir(galleryID), filename)
	_, err := os.Stat(imagePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return Image{}, ErrNotFound
		}
		return Image{}, fmt.Errorf("querying for image: %w", err)
	}

	return Image{
		Filename:  filename,
		GalleryID: galleryID,
		Path:      imagePath,
	}, nil
}

// Delete an image
func (service *GalleryService) DeleteImage(galleryID int, filename string) error {
	image, err := service.Image(galleryID, filename)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	err = os.Remove(image.Path)
	if err != nil {
		return fmt.Errorf("deleting image: %w", err)
	}
	return nil
}

// Upload an image
func (service *GalleryService) CreateImage(galleryID int, filename string, contents io.Reader) error {
	galleryDir := service.galleryDir(galleryID)
	err := os.MkdirAll(galleryDir, 0755)
	/*
		This line creates the galleryDir and all parents directories if they dont exist

		0755 represents the permission code when creating a directory
		0 means it is an octal number
		the second position (7) represents the permissions for Owner. 7 means read, write, execute
		the third position (5) represents the permission for Group. 5 means read, write
		The fourth position (5) represents the permissions for others
		There are more combinations possible. Check online
	*/
	if err != nil {
		return fmt.Errorf("create gallery-%d images directory: %w", galleryID, err)
	}
	imagePath := filepath.Join(galleryDir, filename)
	dst, err := os.Create(imagePath)
	if err != nil {
		return fmt.Errorf("create image file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, contents)
	if err != nil {
		return fmt.Errorf("copying contents to image: %w", err)
	}
	return nil

}
