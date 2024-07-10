package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type Todo struct {
	ID    int
	Title string
	Done  bool
}

type Req struct {
	Id int `form:"id"`
}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	r := gin.Default()
	r.LoadHTMLGlob("./*.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.GET("/todos", getTodoHandler)

	r.POST("/todos", postTodoHandler)

	r.DELETE("/todos/:id", deleteToDoHandler)
	// r.GET("/ping", pingHandler)
	// r.Run(":" + os.Getenv("PORT")) // listen and serve on 0.0.0.0:8080  // cannot use because it not safe close

	srv := http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: r,
	}

	closedChan := make(chan struct{})

	go func() {
		//Done จะ return signal มาก็ต่อเมื่อมี signal ตามที่กำหนดตรงบรรทัด 19
		<-ctx.Done()
		fmt.Println("shutting down....")

		//shutdown เป็นการปิดรับหน้าบ้าน แต่ยังไม่ปิดระบบ
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Println(err)
			}
		}

		close(closedChan)
	}()

	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}

	<-closedChan
	fmt.Println("bye")

}

// func pingHandler(c *gin.Context) {
// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "pong",
// 	})
// }

func save(todos []Todo) {
	data, err := json.Marshal(todos)
	// fmt.Println(todo, data)
	if err != nil {
		// panic(err)
		fmt.Println("Error Kub")
	}

	err = os.WriteFile("todos.json", data, 0644)
	if err != nil {
		// panic(err)
		fmt.Println("Error Kub 2")
	}

}

func postTodoHandler(ctx *gin.Context) {
	todos := readJsonFile()
	var todo Todo

	if err := ctx.ShouldBindJSON(&todo); err != nil {
		ctx.Error(err)
	}

	if len(todos) == 0 {
		todo.ID = 1
	} else {
		todo.ID = len(todos) + 1
	}

	todos = append(todos, todo)
	// fmt.Println(todos)

	save(todos)
	ctx.HTML(http.StatusOK, "todos.html", todos)
	// fmt.Println(todo)
}

func getTodoHandler(ctx *gin.Context) {
	todos := readJsonFile()
	ctx.HTML(http.StatusOK, "todos.html", todos)
}

func readJsonFile() []Todo {
	data, err := os.ReadFile("todos.json")
	if err != nil {
		fmt.Println("Read File Error")
	}

	var todos []Todo

	err = json.Unmarshal(data, &todos)
	if err != nil {
		fmt.Println("Unmarshal json Error")
	}

	return todos
}

func deleteToDoHandler(ctx *gin.Context) {
	// var id Req
	id := ctx.Param("id")
	todos := readJsonFile()

	newTodos := []Todo{}

	for _, e := range todos {
		if strconv.Itoa(e.ID) != id {
			newTodos = append(newTodos, e)
		}
	}
	save(newTodos)
	ctx.HTML(http.StatusOK, "todos.html", newTodos)
}
