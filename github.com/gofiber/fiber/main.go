package main

import (
	"database/sql"
	"flag"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	log "github.com/sirupsen/logrus"
	whatapfiber "github.com/whatap/go-api/instrumentation/github.com/gofiber/fiber"
	"github.com/whatap/go-api/trace"
)

var Database *sql.DB

func index(c *fiber.Ctx) error {
	return c.SendString("Hello, World!")
}

func panicFunc(c *fiber.Ctx) error {
	log.Panic("Panic")
	return c.SendStatus(500)
}

func selectRow(c *fiber.Ctx) error {
	var name string
	row := Database.QueryRow("SELECT name FROM demo_table WHERE id = 1")
	if row == nil {
		return fmt.Errorf("row is nil")
	}
	err := row.Scan(&name)
	if err != nil {
		return err
	}
	c.JSON(map[string]interface{}{
		"id":   "1",
		"name": name,
	})
	return nil
}

func sleepSecond(c *fiber.Ctx) error {
	time.Sleep(time.Second)
	return c.SendString("wake up")
}

func main() {

	portPtr := flag.Int("p", 8081, "web port. default 8081")
	udpPortPtr := flag.Int("up", 6600, "agent port(udp). defalt 6600")
	dataSourcePtr := flag.String("ds", "whatap:whatap1234!@tcp(localhost:3306)/whatap_demo", "dataSourceName")
	flag.Parse()

	port, udpPort, dataSource := *portPtr, *udpPortPtr, *dataSourcePtr

	config := map[string]string{
		"net_udp_port": fmt.Sprintf("%d", udpPort),
		"debug":        "true",
	}
	trace.Init(config)
	defer trace.Shutdown()

	db, err := sql.Open("mysql", dataSource)
	if err != nil {
		log.Panic(err)
	}
	if db == nil {
		log.Panic("Db nil")
	}
	Database = db
	log.Infof("%+v\n", *Database)
	defer Database.Close()

	app := fiber.New(fiber.Config{
		StrictRouting: true,
	})

	app.Use(recover.New())
	app.Use(whatapfiber.Middleware())

	app.Get("/", index)
	app.Get("/panic", panicFunc)
	app.Get("/selectRows", selectRow)
	app.Get("/sleepSecond", sleepSecond)

	log.Fatal(app.Listen(fmt.Sprintf(":%d", port)))
}
