package main

import (
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
)

const MinimumArea = 3000

type MotionDetector struct {
	device     *gocv.VideoCapture
	img        gocv.Mat
	imgDelta   gocv.Mat
	imgThresh  gocv.Mat
	mog2       gocv.BackgroundSubtractorMOG2
	status     string
	color      color.RGBA
	classifier gocv.CascadeClassifier
	writer     *gocv.VideoWriter
}

var blue = color.RGBA{0, 255, 0, 255}

var red = color.RGBA{255, 0, 0, 255}

func NewMotionDetector(deviceID string) (*MotionDetector, error) {
	webcam, err := gocv.OpenVideoCapture(deviceID)
	if err != nil {
		return nil, fmt.Errorf("error opening video capture device: %w", err)
	}

	classifier := gocv.NewCascadeClassifier()
	if !classifier.Load("data/haarcascade_frontalface_default.xml") {
		webcam.Close()
		return nil, fmt.Errorf("error reading cascade file")
	}

	return &MotionDetector{
		device:     webcam,
		img:        gocv.NewMat(),
		imgDelta:   gocv.NewMat(),
		imgThresh:  gocv.NewMat(),
		mog2:       gocv.NewBackgroundSubtractorMOG2(),
		status:     "Ready",
		color:      blue,
		classifier: classifier,
	}, nil
}

func (m *MotionDetector) Close() {
	m.img.Close()
	m.imgDelta.Close()
	m.imgThresh.Close()
	m.mog2.Close()
	m.device.Close()
	m.classifier.Close()
}

func (m *MotionDetector) ProcessFrame() (image.Image, error) {

	if ok := m.device.Read(&m.img); !ok {
		return nil, fmt.Errorf("cannot read from device")
	}

	if m.img.Empty() {
		return nil, fmt.Errorf("empty frame")
	}
	face := m.classifier.DetectMultiScale(m.img)

	m.status = "Ready"
	m.color = color.RGBA{0, 255, 0, 255}

	// get foreground
	m.mog2.Apply(m.img, &m.imgDelta)

	gocv.Threshold(m.imgDelta, &m.imgThresh, 25, 255, gocv.ThresholdBinary)

	// Dilate.
	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Pt(3, 3))
	gocv.Dilate(m.imgThresh, &m.imgThresh, kernel)
	kernel.Close()

	// Find contours.
	contours := gocv.FindContours(m.imgThresh, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	for i := 0; i < contours.Size(); i++ {
		area := gocv.ContourArea(contours.At(i))
		if area < MinimumArea {
			continue
		}

		m.status = "Motion detected"

		// draw contours

		//	m.color = red
		//		gocv.DrawContours(&m.img, contours, i, m.color, 2)

		// Draw bounding rectangle
		rect := gocv.BoundingRect(contours.At(i))
		gocv.Rectangle(&m.img, rect, color.RGBA{0, 0, 255, 255}, 2)
	}

	// status text
	gocv.PutText(&m.img, m.status, image.Pt(10, 30), gocv.FontHersheyPlain, 1.5, m.color, 2)

	for _, r := range face {
		imgFace := m.img.Region(r)
		gocv.Rectangle(&m.img, r, red, 3)
		gocv.GaussianBlur(imgFace, &imgFace, image.Pt(75, 75), 0, 0, gocv.BorderDefault) // blur faces.
		imgFace.Close()
	}

	img, err := m.img.ToImage()
	if err != nil {
		return nil, fmt.Errorf("failed to convert mat to image: %w", err)
	}

	return img, nil
}
