package main

import (
	"github.com/cstegel/opengl-samples-golang/basic-shaders/gfx"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"log"
	"math/rand"
	"runtime"
	"time"
)

const (
	SCREEN_WIDTH  = 230
	SCREEN_HEIGHT = 230
)

var (
	points []float32

	// Set to global for it to be possible to access from keyCallback
	data [SCREEN_WIDTH][SCREEN_HEIGHT]bool
)

func init() {
	// GLFW event handling must be run on the main OS thread
	runtime.LockOSThread()

}

//Resets the data to random values, either to true or false
func randomize(data *[SCREEN_WIDTH][SCREEN_HEIGHT]bool) {
	rand.Seed(int64(time.Now().Second()))
	for i := 0; i < SCREEN_WIDTH; i++ {
		for j := 0; j < SCREEN_HEIGHT; j++ {
			tmp := rand.Intn(2)
			tmpbool := tmp != 0
			data[i][j] = tmpbool
		}
	}
}

// checks if data is set to true and pushes it to the array, which gets rendered
func calculate(data *[SCREEN_WIDTH][SCREEN_HEIGHT]bool) {
	points = nil
	for i := 0; i < SCREEN_WIDTH; i++ {
		for j := 0; j < SCREEN_HEIGHT; j++ {
			if (data[i][j]) == true {

				var xpos float32
				var ypos float32

				// Opengl coordinates range from -1 to 1, 0 being center
				xpos = (float32(i) * 2 / float32(SCREEN_WIDTH)) - 1
				ypos = (float32(j) * 2 / float32(SCREEN_HEIGHT)) - 1

				points = append(points, xpos, ypos, 0)

			}
		}
	}
}

// Calculates total number alive of neighbours
func check_neigbours(xcoord, ycoord int, newcycle [SCREEN_WIDTH][SCREEN_HEIGHT]bool) int {

	sum := 0
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {

			col := (xcoord + i + SCREEN_HEIGHT) % SCREEN_HEIGHT
			row := (ycoord + j + SCREEN_WIDTH) % SCREEN_WIDTH

			if newcycle[row][col] {
				sum++
			}
		}
	}
	if newcycle[xcoord][ycoord] {
		sum--
	}

	return sum
}

// A new cell is born when there are exactly 3 alive cells nearby
// A cell dies when there are less than two 2 or more than 3 alive

// Pushes the information for the next frame
func cycle(data [SCREEN_WIDTH][SCREEN_HEIGHT]bool) [SCREEN_WIDTH][SCREEN_HEIGHT]bool {
	var newcycle [SCREEN_WIDTH][SCREEN_HEIGHT]bool
	for i := 0; i < SCREEN_WIDTH; i++ {
		for j := 0; j < SCREEN_HEIGHT; j++ {

			count := check_neigbours(i, j, data)
			if (data[i][j]) == false && count == 3 {
				newcycle[i][j] = true
			} else if (data[i][j]) == true && (count < 2 || count > 3) {
				newcycle[i][j] = false
			} else {
				newcycle[i][j] = data[i][j]
			}
		}
	}

	return newcycle
}

func main() {

	// Fill the array with random 0-1 values
	randomize(&data)

	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to inifitialize glfw:", err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(SCREEN_WIDTH, SCREEN_HEIGHT, "Conway's Game of Life", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// Initialize Glow (go function bindings)
	if err := gl.Init(); err != nil {
		panic(err)
	}

	window.SetKeyCallback(keyCallback)

	err = mainloop(window)
	if err != nil {
		log.Fatal(err)
	}
}

func vaogen(points []float32) uint32 {

	var vbo uint32
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(points), gl.Ptr(points), gl.STATIC_DRAW)

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)
	gl.EnableVertexAttribArray(0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 0, nil)

	return vao
}

func mainloop(window *glfw.Window) error {

	// the linked shader program determines how the data will be rendered
	vertShader, err := gfx.NewShaderFromFile("shaders/basic.vert", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragShader, err := gfx.NewShaderFromFile("shaders/basic.frag", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	shaderProgram, err := gfx.NewProgram(vertShader, fragShader)
	if err != nil {
		return err
	}
	defer shaderProgram.Delete()

	for !window.ShouldClose() {

		// poll events and call their registered callbacks
		glfw.PollEvents()

		// clear buffer data
		gl.ClearColor(0, 0, 0, 1.0)
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		// add position of pixels to the array
		calculate(&data)

		//Get new life-cycle
		data = cycle(data)

		// perform rendering
		shaderProgram.Use()
		gl.BindVertexArray(vaogen(points))
		gl.DrawArrays(gl.POINTS, 0, int32(len(points)/3))

		window.SwapBuffers()
	}

	return nil
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action,
	mods glfw.ModifierKey) {

	// When a user presses the escape key, we set the WindowShouldClose property to true,
	// which closes the application
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}

	if key == glfw.KeySpace && action == glfw.Press {
		randomize(&data)
	}

}
