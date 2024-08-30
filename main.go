package main

func main() {
	app := NewApp()
	app.BindEvents()
	app.Render()
}
