package endpoint

import (
	"net/http"

	"0lvl/internal/repository"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog"
)

var (
	msgNoData = []byte(`{"message": "No data"}`)
)

type Endpoint struct {
	repo *repository.Repo
	log  zerolog.Logger
}

func New(repo *repository.Repo, log zerolog.Logger) *Endpoint {
	return &Endpoint{
		repo: repo,
		log:  log,
	}
}

func (e *Endpoint) Run() {
	router := httprouter.New()
	router.GET("/", e.index)
	router.GET("/order/:uid", e.order)
	router.GET("/metric", e.metrica)

	server := &http.Server{
		Addr:    ":8000",
		Handler: router,
	}
	err := server.ListenAndServe()
	if err != nil {
		e.log.Fatal().Err(err).Msg("fail listen")
	}
}

func (e *Endpoint) index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b := e.repo.OrdersLink(32)
	w.Write(b)
}

func (e *Endpoint) order(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	uid := ps.ByName("uid")
	b, err := e.repo.Order(uid)
	if err != nil {
		e.log.Err(err).Msg("")
		w.WriteHeader(404)
		w.Write(msgNoData)
		return
	}
	w.Write(b)
}

func (e *Endpoint) metrica(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	b := e.repo.Metrica()
	w.Write(b)
}
