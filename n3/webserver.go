// webserver.go

package n3

import (
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func RunWebserver(webPort int, localBlockchain *Blockchain) {

	// create stan connection for writing to feed
	sc, err := NSSConnection("n3web")
	if err != nil {
		log.Println("cannot connect to nss: ", err)
		return
	}
	defer sc.Close()

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Route => handler
	e.POST("/tuple", func(c echo.Context) error {

		// unpack tuple from paylaod
		t := new(SPOTuple)
		if err = c.Bind(t); err != nil {
			return err
		}

		// add to the blockchain
		b, err := localBlockchain.AddNewBlock(t)
		if err != nil {
			log.Println("error adding data block via web:", err)
			return err
		}

		blockBytes := b.Serialize()
		err = sc.Publish("feed", blockBytes)
		if err != nil {
			log.Println("web handler cannot send new block to feed: ", err)
			return err
		}

		return c.JSON(http.StatusOK, t)

	})

	// Start server
	addr := fmt.Sprintf(":%d", webPort)
	e.Logger.Fatal(e.Start(addr))
}
