package cloudinary

import (
	"context"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// Client wraps the Cloudinary SDK client.
type Client struct {
	cld    *cloudinary.Cloudinary
	folder string
}

// NewClient initialises a Cloudinary client from credentials.
// cloudName, apiKey, apiSecret come from your Cloudinary dashboard.
func NewClient(cloudName, apiKey, apiSecret, folder string) (*Client, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}
	return &Client{cld: cld, folder: folder}, nil
}

// UploadProfilePicture uploads a multipart file to Cloudinary and returns the secure URL.
func (c *Client) UploadProfilePicture(ctx context.Context, file multipart.File, filename string) (string, error) {
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:         c.folder,
		PublicID:       filename,
		UniqueFilename: ptrBool(true),
		Overwrite:      ptrBool(false),
	})
	if err != nil {
		return "", err
	}
	return resp.SecureURL, nil
}

func ptrBool(b bool) *bool {
	return &b
}
