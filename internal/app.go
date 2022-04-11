package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/stan.go"

	migration "wildber/db"
	"wildber/internal/config"
	"wildber/internal/model"
	"wildber/internal/repository"
)

type app struct {
	cfg       *config.Config
	ctx       context.Context
	orderRepo *repository.OrderRepo
}

func NewApp(cfg *config.Config, pool *pgxpool.Pool, ctx context.Context) (App, error) {

	orderRepo, err := repository.NewOrderRepo(pool)
	if err != nil {
		log.Fatal(err)
	}

	return &app{
		cfg:       cfg,
		ctx:       ctx,
		orderRepo: orderRepo,
	}, nil
}

func (a *app) Run() {
	a.postgresMigration()
	go a.startHttpserver()
	a.startNatService()
}

func (a *app) postgresMigration() {
	err := migration.MigrateUp("postgres", a.cfg.Postgres.Url)
	if err != nil {
		log.Fatal(err)
	}
}

func (a *app) startNatService() {
	clusterID := "test-cluster"
	clientID := "stan-sub"
	subject := "wildber"
	var durable string
	unsubscribe := false

	sc, err := stan.Connect(clusterID, clientID)
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at:", err)
	}
	log.Printf("Connected to %s clusterID: [%s] clientID: [%s]\n", clusterID, clientID)

	aw, _ := time.ParseDuration("30s")
	sub, err := sc.Subscribe(subject, func(msg *stan.Msg) {
		msg.Ack()
		var m model.MessageJson
		err := json.Unmarshal(msg.Data, &m)
		if err != nil {
			log.Println(err)
			return
		}
		id, _ := a.orderRepo.InsertOrder(a.ctx, m)
		log.Printf("[%d]Received: %s ", id, m)
	}, stan.SetManualAckMode(), stan.AckWait(aw))

	if err != nil {
		sc.Close()
		log.Fatal(err)
	}

	log.Printf("Listening on [%s], clientID=[%s]\n", subject, clientID)

	// Wait for a SIGINT (perhaps triggered by user with CTRL-C)
	// Run cleanup when signal is received
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for range signalChan {
			fmt.Printf("\nReceived an interrupt, unsubscribing and closing connection...\n\n")
			// Do not unsubscribe a durable on exit, except if asked to.
			if durable == "" || unsubscribe {
				sub.Unsubscribe()
			}
			sc.Close()
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}

func (a *app) startHttpserver() {
	e := echo.New()

	// Routes
	e.Static("/assets", "internal/view/assets")
	e.GET("/order/:id", func(c echo.Context) error {
		id, _ := strconv.Atoi(c.Param("id"))
		res, err := a.orderRepo.GetOrderById(a.ctx, id)
		if err != nil {
			return c.String(http.StatusNotFound, "{}")
		}
		return c.JSON(http.StatusOK, res)
	})

	e.GET("/", func(c echo.Context) error {
		filename, err := os.Open("internal/view/home.html")
		if err != nil {
			log.Fatal(err)
		}
		defer filename.Close()
		data, err := ioutil.ReadAll(filename)
		if err != nil {
			log.Fatal(err)
		}

		return c.HTML(http.StatusOK, string(data))
	})

	// Start server
	e.Logger.Fatal(e.Start(":8081"))
}
