package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/basicfont"
)

type Game struct {
	detector      *MotionDetector
	webcamTexture *ebiten.Image
	width         int
	height        int
}

func (g *Game) Update() error {
	img, err := g.detector.ProcessFrame()
	if err != nil {
		log.Printf("Error processing frame: %v", err)
		return nil
	}

	// Initialize texture on first frame
	if g.webcamTexture == nil {
		bounds := img.Bounds()
		g.width = bounds.Dx()
		g.height = bounds.Dy()
		g.webcamTexture = ebiten.NewImage(g.width, g.height)
	}

	// Update texture
	g.webcamTexture.WritePixels(imageToBytes(img))

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.webcamTexture != nil {
		screen.DrawImage(g.webcamTexture, nil)
	}

	// Draw status overlay
	statusText := fmt.Sprintf("FPS: %.1f | Status: %s", ebiten.ActualFPS(), g.detector.status)
	text.Draw(screen, statusText, basicfont.Face7x13, 10, g.height-10, color.White)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.width > 0 && g.height > 0 {
		return g.width, g.height
	}
	return 640, 480
}

// Convert image to byte slice for WritePixels (RGBA format)
func imageToBytes(img image.Image) []byte {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	pixels := make([]byte, width*height*4)

	i := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			pixels[i] = byte(r >> 8)
			pixels[i+1] = byte(g >> 8)
			pixels[i+2] = byte(b >> 8)
			pixels[i+3] = byte(a >> 8)
			i += 4
		}
	}
	return pixels
}

func main() {
	deviceID := "0"
	if len(os.Args) >= 2 {
		deviceID = os.Args[1]
	}

	detector, err := NewMotionDetector(deviceID)
	if err != nil {
		log.Fatal(err)
	}
	defer detector.Close()

	fmt.Printf("Start reading device: %v\n", deviceID)

	game := &Game{
		detector: detector,
	}

	ebiten.SetWindowTitle("MT")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	// Run game loop
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
