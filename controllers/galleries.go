package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"github.com/TanmayKhot/pixvault/context"
	"github.com/TanmayKhot/pixvault/errors"
	"github.com/TanmayKhot/pixvault/models"
	"github.com/go-chi/chi/v5"
)

// rendering the new gallery page
type Galleries struct {
	Templates struct {
		New   Template
		Edit  Template
		Index Template
		Show  Template
	}
	GalleryService *models.GalleryService
}

// This is a type of function that takes the given inputs and returns an error
// Any function that takes in these inputs and returns an error is of type galleryOpt
// The primary purpose of this type of functions is to perform middleware actions on Gallery
// such as authenticating user
type galleryOpt func(http.ResponseWriter, *http.Request, *models.Gallery) error

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")
	g.Templates.New.Execute(w, r, data)
}

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data struct {
		UserID int
		Title  string
	}
	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")

	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		g.Templates.New.Execute(w, r, data, err)
		return
	}

	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

// The 'opts ...galleryOpt' is a parameter which accepts any numer of galleryOpt type of functions (defined above)
// Using that allows us to pass functions as an input to galleryByID
// That way, we can authenticate the user access right after fetching the gallery by ID
// Or we can pass other middleware applications which are of type galleryOpt which perform some action on the gallery and return an error if possible
func (g Galleries) galleryByID(w http.ResponseWriter, r *http.Request, opts ...galleryOpt) (*models.Gallery, error) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return nil, err
	}
	gallery, err := g.GalleryService.ByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return nil, err
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return nil, err
	}

	// 'range opts' will return (index, function)
	// We don't care about the index so we skip that
	// All functions are used, parameters are passed and if there's an error, we return that
	// Currently we only have user authentication, but if we develop an other such method, we can pass it here
	// as a variadic parameter and it will be evaluated in this for loop
	for _, opt := range opts {
		err = opt(w, r, gallery)
		if err != nil {
			return nil, err
		}
	}
	return gallery, nil
}

// userMustOwnGallery is of type galleryOpt because it has the same input and output types
func userMustOwnGallery(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	user := context.User(r.Context())

	if user == nil || user.ID != gallery.UserID {
		http.Error(w, "You are not authorized to view this gallery", http.StatusForbidden)
		return fmt.Errorf("user does not have access to this gallery")
	}
	return nil
}

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnGallery) // Pass userMustOwnGallery as a variaidc parameter
	if err != nil {
		return
	}

	// Before we can render the gallery information, we need to verify that the user actually owns this gallery.
	err = userMustOwnGallery(w, r, gallery)
	if err != nil {
		return
	}

	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}

	data := struct {
		ID     int
		Title  string
		Access string
		Images []Image
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	data.Access = "Public"
	if gallery.IsPrivate {
		data.Access = "Private"
	}

	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}
	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnGallery) // Pass userMustOwnGallery as a variaidc parameter
	if err != nil {
		return
	}
	// Before we can render the gallery information, we need to verify that the user actually owns this gallery.
	err = userMustOwnGallery(w, r, gallery)
	if err != nil {
		return
	}

	title := r.FormValue("title")
	gallery.Title = title
	err = g.GalleryService.Update(gallery)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) ChangeAccess(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = g.GalleryService.ChangeAccess(gallery.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID     int
		Title  string
		Access string
	}
	var data struct {
		Galleries []Gallery
	}

	user := context.User(r.Context())
	galleries, err := g.GalleryService.ByUserID(user.ID)
	if err != nil {
		fmt.Printf("Type of user.ID: %T\n", user.ID)
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	for _, gallery := range galleries {
		access := "Public"
		if gallery.IsPrivate {
			access = "Private"
		}
		data.Galleries = append(data.Galleries, Gallery{
			ID:     gallery.ID,
			Title:  gallery.Title,
			Access: access,
		})

	}

	g.Templates.Index.Execute(w, r, data)
}

// Anyone with a link to a gallery will be able to view it, even users who have not signed in
func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r)
	if err != nil {
		return
	}

	if gallery.IsPrivate {

		err = userMustOwnGallery(w, r, gallery)
		if err != nil {
			return
		}

	}

	// Prepare image struct to show single image
	type Image struct {
		GalleryID       int
		Filename        string
		FilenameEscaped string
	}

	// Prepare data to show all images of gallery
	var data struct {
		ID        int
		Title     string
		Images    []Image
		UserID    int
		UserEmail string
		Access    string
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	data.UserID = gallery.UserID
	data.UserEmail = gallery.UserEmail
	data.Access = "Public"
	if gallery.IsPrivate {
		data.Access = "Private"
	}

	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	/*
		Display dummy cat pictures as placeholders

		for i := 0; i < 20; i++ {
			// width and height are random values betwee 200 and 700
			w, h := rand.Intn(500)+200, rand.Intn(500)+200
			// using the width and height, we generate a URL
			catImageURL := fmt.Sprintf("https://placecats.com/%d/%d", w, h)
			// Then we add the URL to our images.
			data.Images = append(data.Images, catImageURL)
		}
	*/

	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename), // using PathEscape to include characters for file names which could be URL restricted
		})
	}

	g.Templates.Show.Execute(w, r, data)
}

func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	err = g.GalleryService.Delete(gallery.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g Galleries) filename(w http.ResponseWriter, r *http.Request) string {
	filename := chi.URLParam(r, "filename")
	filename = filepath.Base(filename)
	return filename
}

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}

	images, err := g.GalleryService.Images(galleryID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	var requestedImage models.Image
	imageFound := false
	for _, image := range images {
		if filename == image.Filename {
			requestedImage = image
			imageFound = true
			break
		}
	}

	if !imageFound {
		http.Error(w, "Image not found", http.StatusNotFound)
	}

	http.ServeFile(w, r, requestedImage.Path)
}

/*

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	filename := chi.URLParam(r, "filename")
	galleryID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}
	image, err := g.GalleryService.Image(galleryID, filename)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, image.Path)
}
*/

func (g Galleries) DeleteImage(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = g.GalleryService.DeleteImage(gallery.ID, filename)
	if err != nil {
		http.Error(w, "something went wrong: ", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) UploadImages(w http.ResponseWriter, r *http.Request) {
	gallery, err := g.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}
	err = r.ParseMultipartForm(5 << 20) // 5 MB = 5 x 2^20
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	fileHeaders := r.MultipartForm.File["images"]
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		err = g.GalleryService.CreateImage(gallery.ID, fileHeader.Filename, file)
		if err != nil {
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
	}
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusFound)
}
