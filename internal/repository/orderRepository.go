package repository

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"gopkg.in/errgo.v2/fmt/errors"

	"wildber/internal/cache"
	"wildber/internal/model"
)

type OrderRepo struct {
	pool  *pgxpool.Pool
	cache *cache.Cache
}

func NewOrderRepo(pool *pgxpool.Pool) (*OrderRepo, error) {
	cache := cache.New(5*time.Minute, 10*time.Minute)

	return &OrderRepo{
		pool:  pool,
		cache: cache,
	}, nil
}

func (s *OrderRepo) InsertOrder(ctx context.Context, jsonData model.MessageJson) (int, error) {

	data, err := json.Marshal(jsonData)
	if err != nil {
		log.Println(err)
	}
	var id int
	if err := s.pool.QueryRow(ctx, `INSERT INTO public.order(data) VALUES ($1) RETURNING id`, data).Scan(&id); err != nil {
		log.Printf("Order %d \n", err)
		return 0, err
	}
	s.cache.Set(string(id), string(data), 5*time.Minute)
	return id, nil
}

func (s *OrderRepo) GetOrderById(ctx context.Context, id int) (string, error) {

	log.Println(id)
	order, ok := s.cache.Get(string(id))
	if ok == true {
		if order == "" {
			return order, errors.Newf("No data '%s'", order)
		}
		return order, nil
	}

	log.Print("ПОЛУЧАЕМ ИЗ БД И ЗАПИСЫВАЕМ В КЕШ", id)
	var data string
	if err := s.pool.QueryRow(ctx, "select data from public.order where id= $1", id).Scan(&data); err != nil {
		log.Printf("Order %d,    %d \n", data, err)
		s.cache.Set(string(id), data, 5*time.Minute)
		return data, errors.Newf("No data '%s'", data)
	}
	s.cache.Set(string(id), data, 5*time.Minute)
	if data == "" {
		return data, errors.Newf("No data '%s'", data)
	}
	return data, nil
}
