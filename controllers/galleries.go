package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
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
	if user.ID != gallery.UserID {
		http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
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

	data := struct {
		ID    int
		Title string
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
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
	editPath := fmt.Sprintf("/galleries/%d/edit", gallery.ID)
	http.Redirect(w, r, editPath, http.StatusNotFound)
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID    int
		Title string
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
		data.Galleries = append(data.Galleries, Gallery{
			ID:    gallery.ID,
			Title: gallery.Title,
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

	var data struct {
		ID     int
		Title  string
		Images []string
	}
	data.ID = gallery.ID
	data.Title = gallery.Title

	for i := 0; i < 20; i++ {
		// width and height are random values betwee 200 and 700
		w, h := rand.Intn(500)+200, rand.Intn(500)+200
		// using the width and height, we generate a URL
		catImageURL := fmt.Sprintf("https://placecats.com/%d/%d", w, h)
		// Then we add the URL to our images.
		data.Images = append(data.Images, catImageURL)
	}

	g.Templates.Show.Execute(w, r, data)
}
