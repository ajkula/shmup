package graphics

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Point3D struct {
	X, Y, Z float64
}

func DrawTriangle(screen *ebiten.Image, x, y, size float64, rotation float64, c color.Color) {
	// Define triangle points
	points := []struct{ x, y float64 }{
		{0, size / 2},          // Top
		{-size / 2, -size / 2}, // Bottom right
		{size / 2, -size / 2},  // Bottom left
	}

	// Rotate and translate points
	cosR, sinR := math.Cos(rotation), math.Sin(rotation)
	for i := range points {
		px, py := points[i].x, points[i].y
		points[i].x = px*cosR - py*sinR + x
		points[i].y = px*sinR + py*cosR + y
	}

	// Draw the triangle
	vector.StrokeLine(screen, float32(points[0].x), float32(points[0].y), float32(points[1].x), float32(points[1].y), 1, c, true)
	vector.StrokeLine(screen, float32(points[1].x), float32(points[1].y), float32(points[2].x), float32(points[2].y), 1, c, true)
	vector.StrokeLine(screen, float32(points[2].x), float32(points[2].y), float32(points[0].x), float32(points[0].y), 1, c, true)
}

func DrawCube3D(screen *ebiten.Image, x, y, z, size float64, c color.Color) {
	// Isometric projection constants
	cos30 := math.Cos(math.Pi / 6)
	sin30 := math.Sin(math.Pi / 6)

	// Calculate cube vertices
	halfSize := size / 2
	vertices := []struct{ x, y, z float64 }{
		{-halfSize, -halfSize, -halfSize},
		{halfSize, -halfSize, -halfSize},
		{halfSize, halfSize, -halfSize},
		{-halfSize, halfSize, -halfSize},
		{-halfSize, -halfSize, halfSize},
		{halfSize, -halfSize, halfSize},
		{halfSize, halfSize, halfSize},
		{-halfSize, halfSize, halfSize},
	}

	// Project 3D points to 2D isometric view
	points2D := make([][2]float64, 8)
	for i, v := range vertices {
		isoX := (v.x - v.y) * cos30
		isoY := (v.x+v.y)*sin30 - v.z
		points2D[i][0] = x + isoX
		points2D[i][1] = y + isoY
	}

	// Define edges
	edges := [][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, // Bottom face
		{4, 5}, {5, 6}, {6, 7}, {7, 4}, // Top face
		{0, 4}, {1, 5}, {2, 6}, {3, 7}, // Connecting edges
	}

	// Draw edges
	for _, edge := range edges {
		p1 := points2D[edge[0]]
		p2 := points2D[edge[1]]
		vector.StrokeLine(screen, float32(p1[0]), float32(p1[1]), float32(p2[0]), float32(p2[1]), 1, c, true)
	}
}

func DrawCircle(screen *ebiten.Image, x, y, size float64, c color.Color) {
	vector.DrawFilledCircle(screen, float32(x+size/2), float32(y+size/2), float32(size/2), c, true)
}

func DrawCube(screen *ebiten.Image, x, y, width, height, angle float64, c color.Color) {
	vertices := []struct{ x, y float64 }{
		{-1, -1}, {1, -1}, {1, 1}, {-1, 1},
		{-1, -1}, {1, -1}, {1, 1}, {-1, 1},
	}

	for i, v := range vertices[:4] {
		vertices[i].x = v.x*math.Cos(angle) - v.y*math.Sin(angle)
		vertices[i].y = v.x*math.Sin(angle) + v.y*math.Cos(angle)
	}

	for i := 0; i < 4; i++ {
		x1, y1 := vertices[i].x*width+x, vertices[i].y*width+y
		x2, y2 := vertices[(i+1)%4].x*height+x, vertices[(i+1)%4].y*height+y
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 1, c, false)
		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x1), float32(y1-width), 1, c, false)
	}
}
