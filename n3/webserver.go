// webserver.go

package n3

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func RunWebserver(webPort int, hexastore *Hexastore) {

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
	// TODO parameterise context
	e.POST("/tuple", func(c echo.Context) error {

		// unpack tuple from payload
		t := new(SPOTuple)
		if err = c.Bind(t); err != nil {
			return err
		}

		// add to the blockchain
		localBlockchain := GetBlockchain("SIF", cs.PublicID())
		b, err := localBlockchain.AddNewBlock(t)
		if err != nil {
			log.Println("error adding data block via web:", err)
			return err
		}

		err = sc.Publish("feed", b.Serialize())
		if err != nil {
			log.Println("web handler cannot send new block to feed: ", err)
			return err
		}

		return c.JSON(http.StatusOK, t)

	})

	// TODO parameterise context
	e.POST("/tuples", func(c echo.Context) error {

		// unpack tuple from payload
		tuples := make([]SPOTuple, 0)
		if err = c.Bind(&tuples); err != nil {
			log.Println(err)
			return err
		}
		//log.Printf("/tuples: %#v\n", tuples)

		// add to the blockchain
		localBlockchain := GetBlockchain("SIF", cs.PublicID())
		for _, t := range tuples {
			b, err := localBlockchain.AddNewBlock(&t)
			if err != nil {
				log.Println("error adding data block via web:", err)
				return err
			}

			err = sc.Publish("feed", b.Serialize())
			if err != nil {
				log.Println("web handler cannot send new block to feed: ", err)
				return err
			}
		}

		return c.JSON(http.StatusOK, tuples)

	})

	e.GET("/HasKey/:key", func(c echo.Context) error {
		key := c.Param("key")
		//log.Printf("Has key: %s\n", key)
		has, err := hexastore.HasKey(key)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		if has {
			c.String(http.StatusOK, "true")
		} else {
			c.String(http.StatusNotFound, "false")
		}
		return nil
	})

	e.GET("/tuple/:key", func(c echo.Context) error {
		key := c.Param("key")
		tuples, err := hexastore.GetTuples(key)
		//log.Printf("%+v\n", tuples)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.Response().Header().Set("Content-Type", "application/json")
		ret, err := json.Marshal(tuples)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.String(http.StatusOK, string(ret))
		return nil
	})

	/* NSW DIG hardcoded queries */
	e.GET("/kla2student", func(c echo.Context) error {
		kla := c.QueryParam("kla")
		yrlvl := c.QueryParam("yrlvl")
		ids, err := hexastore.KLAtoStudentQuery(kla, yrlvl)
		//log.Printf("kla2student: %+v\n", ids)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.Response().Header().Set("Content-Type", "application/json")
		ret, err := json.Marshal(ids)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.String(http.StatusOK, string(ret))
		return nil
	})

	e.GET("/kla2staff", func(c echo.Context) error {
		kla := c.QueryParam("kla")
		yrlvl := c.QueryParam("yrlvl")
		ids, err := hexastore.KLAtoTeacherQuery(kla, yrlvl)
		//log.Printf("%+v\n", tuples)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.Response().Header().Set("Content-Type", "application/json")
		ret, err := json.Marshal(ids)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.String(http.StatusOK, string(ret))
		return nil
	})

	e.GET("/kla2teachinggroup", func(c echo.Context) error {
		kla := c.QueryParam("kla")
		yrlvl := c.QueryParam("yrlvl")
		ids, err := hexastore.KLAtoTeachingGroupQuery(kla, yrlvl)
		//log.Printf("%+v\n", tuples)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.Response().Header().Set("Content-Type", "application/json")
		ret, err := json.Marshal(ids)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.String(http.StatusOK, string(ret))
		return nil
	})

	e.GET("/kla2timetablesubject", func(c echo.Context) error {
		kla := c.QueryParam("kla")
		yrlvl := c.QueryParam("yrlvl")
		ids, err := hexastore.KLAtoTimeTableSubjectQuery(kla, yrlvl)
		//log.Printf("%+v\n", ids)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.Response().Header().Set("Content-Type", "application/json")
		ret, err := json.Marshal(ids)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return err
		}
		c.String(http.StatusOK, string(ret))
		return nil
	})

	// Start server
	addr := fmt.Sprintf(":%d", webPort)
	e.Logger.Fatal(e.Start(addr))
}
