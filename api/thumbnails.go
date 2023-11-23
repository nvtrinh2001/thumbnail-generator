package handler

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"net/http"
	"os"
	"strings"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	godraw "golang.org/x/image/draw"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomedium"
	"golang.org/x/image/font/gofont/goregular"
)

func generateThumbnail(topic, title string) image.Image {
	thumbnailWidth := 1200
	thumbnailHeight := 630

	img := image.NewRGBA(image.Rect(0, 0, thumbnailWidth, thumbnailHeight))

	// Fill the image with a background color
	backgroundColor := color.RGBA{R: 23, G: 23, B: 23, A: 255} // White color
	draw.Draw(img, img.Bounds(), &image.Uniform{C: backgroundColor}, image.Point{}, draw.Src)

	// Draw topic on the thumbnail
	topicFontSize := 40.0
	topicColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // Black color
	drawTopic(img, topic, topicFontSize, topicColor)

	// Draw title on the thumbnail
	titleFontSize := 60.0
	titleColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // Gray color
	drawTitle(img, title, titleFontSize, titleColor)

	// Draw author on the thumbnail
	authorFontSize := 40.0
	authorColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // Gray color
	drawAuthor(img, "piklr", authorFontSize, authorColor)

	return img
}

func drawText(img *image.RGBA, text string, fontSize float64, textColor color.RGBA, posX, posY int, fnt *truetype.Font) {
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(fnt)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(textColor))

	pt := freetype.Pt(posX, posY+int(c.PointToFixed(fontSize)>>6))
	_, err := c.DrawString(text, pt)
	if err != nil {
		fmt.Println("Error drawing text:", err)
	}
}

func drawTopic(img *image.RGBA, topic string, fontSize float64, textColor color.RGBA) {
	topicPosX := 80
	topicPosY := 80

	// Use Go Regular font
	fnt := loadFont(gomedium.TTF)

	drawText(img, topic, fontSize, textColor, topicPosX, topicPosY, fnt)
}

// func drawTitle(img *image.RGBA, title string, fontSize float64, textColor color.RGBA) {
// 	titlePosX := 80
// 	titlePosY := 280
//
// 	// Use Go Regular font
// 	fnt := loadFont(gobold.TTF)
//
// 	drawText(img, title, fontSize, textColor, titlePosX, titlePosY, fnt)
// }

func drawAuthor(img *image.RGBA, title string, fontSize float64, textColor color.RGBA) {
	titlePosX := 80
	titlePosY := 480

	// Use Go Regular font
	fnt := loadFont(goregular.TTF)

	drawText(img, title, fontSize, textColor, titlePosX, titlePosY, fnt)
}

func loadFont(fontBytes []byte) *truetype.Font {
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println("Error loading font:", err)
		return nil
	}
	return font
}

type circle struct {
	p image.Point
	r int
}

func (c *circle) ColorModel() color.Model {
	return color.AlphaModel
}

func (c *circle) Bounds() image.Rectangle {
	return image.Rect(c.p.X-c.r, c.p.Y-c.r, c.p.X+c.r, c.p.Y+c.r)
}

func (c *circle) At(x, y int) color.Color {
	xx, yy, rr := float64(x-c.p.X)+0.5, float64(y-c.p.Y)+0.5, float64(c.r)
	if xx*xx+yy*yy < rr*rr {
		return color.Alpha{255}
	}
	return color.Alpha{0}
}

func Thumbnails(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	title := r.URL.Query().Get("title")

	thumbnail := generateThumbnail(topic, title)

	externalImgFile, err := os.Open("ghost-2.jpg")
	if err != nil {
		fmt.Println("Error opening base image:", err)
		return
	}
	defer externalImgFile.Close()

	externalImg, _, err := image.Decode(externalImgFile)
	if err != nil {
		fmt.Println("Error decoding base image:", err)
		return
	}

	newImg := image.NewRGBA(thumbnail.Bounds())

	// Set the expected size that you want:
	resizedImg := image.NewRGBA(image.Rect(0, 0, 160, 160))

	// Resize:
	godraw.NearestNeighbor.Scale(resizedImg, resizedImg.Rect, externalImg, externalImg.Bounds(), godraw.Over, nil)

	// Draw the base image onto the new image
	draw.Draw(newImg, thumbnail.Bounds(), thumbnail, image.Point{}, draw.Src)

	// Calculate the position to place the external image on the base image
	externalImgPosition := image.Point{X: newImg.Rect.Dx() - resizedImg.Rect.Dx() - 80, Y: newImg.Rect.Dy() - resizedImg.Rect.Dy() - 80}

	maskedImg := image.NewRGBA(resizedImg.Bounds())

	draw.DrawMask(maskedImg, maskedImg.Bounds(), resizedImg, image.Point{}, &circle{image.Point{80, 80}, 80}, image.Point{}, draw.Over)

	// Draw the external image onto the new image at the specified position
	draw.Draw(newImg, resizedImg.Bounds().Add(externalImgPosition), maskedImg, image.Point{}, draw.Over)

	// Save the resulting image to a file
	outputFile, err := os.Create("output-image.jpg") // Replace with your desired output file path
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Encode the new image and save it to the output file
	err = png.Encode(outputFile, newImg)
	if err != nil {
		fmt.Println("Error encoding image:", err)
		return
	}

	w.Header().Set("Content-Type", "image/png")

	err = png.Encode(w, newImg)
	if err != nil {
		http.Error(w, "Error encoding image", http.StatusInternalServerError)
		return
	}
}

func drawTitle(img *image.RGBA, title string, fontSize float64, textColor color.RGBA) {
	titlePosX := 80
	titlePosY := 240
	maxWidth := 2000 // Maximum width for wrapping

	// Use Go Bold font
	fnt := loadFont(gobold.TTF)

	// Create a new freetype context for drawing text
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(fnt)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(textColor))

	// Define the wrapping function for the title text
	drawWrappedText(c, title, fontSize, textColor, titlePosX, titlePosY, maxWidth)
}

func drawWrappedText(c *freetype.Context, text string, fontSize float64, textColor color.RGBA, posX, posY, maxWidth int) {
	words := strings.Fields(text)
	line := ""
	x := posX
	y := posY

	for _, word := range words {
		measure := c.PointToFixed(fontSize * float64(len(line+word)))
		if x+int(measure)>>6 > maxWidth {
			drawTextPart(c, line, fontSize, textColor, x, y, posX)
			y += int(fontSize * 1.5)
			line = ""
		}
		line += word + " "
	}

	// Draw the remaining text
	drawTextPart(c, line, fontSize, textColor, x, y, posX)
}

func drawTextPart(c *freetype.Context, text string, fontSize float64, textColor color.RGBA, x, y, posX int) {
	pt := freetype.Pt(x, y+int(c.PointToFixed(fontSize)>>6))
	_, err := c.DrawString(text, pt)
	if err != nil {
		fmt.Println("Error drawing text:", err)
	}
}
