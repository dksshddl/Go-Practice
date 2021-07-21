package main

import (
	"Go-Practice/api/respond"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type poll struct {
	ID      primitive.ObjectID `bson:"_id" json:"id"`
	Title   string             `json:"title"`
	Options []string           `json:"options"`
	Results map[string]int     `json:"results,omitempty"`
	APIKey  string             `json:"apikey"`
}

func (s *Server) handlePolls(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		s.handlePollsGet(w, r)
		return
	case "POST":
		s.handlePollPost(w, r)
		return
	case "DELETE":
		s.handlePollsDelete(w, r)
		return
	case "OPTIONS":
		w.Header().Add("Access-Control-Allow-Methods", "DELETE")
		respond.Respond(w, r, http.StatusOK, nil)
		return
	}
	respond.RespondHTTPErr(w, r, http.StatusNotFound)
}

func (s *Server) handlePollsGet(w http.ResponseWriter, r *http.Request) {
	session, err := s.db.StartSession()
	if err != nil {
		log.Fatalln(err)
	}

	defer session.EndSession(r.Context())
	c := session.Client().Database("ballots").Collection("polls")
	var cursor *mongo.Cursor
	p := NewPath(r.URL.Path)
	if p.HasID() {
		log.Println("detect id, ", p.Id)
		objectId, err := primitive.ObjectIDFromHex(p.Id)
		if err != nil {
			log.Fatalln(err)
		}
		cursor, _ = c.Find(r.Context(), bson.M{"_id": objectId})
	} else {
		cursor, _ = c.Find(r.Context(), bson.M{})
	}
	var result []*poll
	if err := cursor.All(r.Context(), &result); err != nil {
		respond.RespondErr(w, r, http.StatusInternalServerError, err)
		return
	}
	respond.Respond(w, r, http.StatusOK, &result)
}

func (s *Server) handlePollPost(w http.ResponseWriter, r *http.Request) {
	session, err := s.db.StartSession()
	if err != nil {
		log.Fatalln(err)
	}
	defer session.EndSession(r.Context())
	c := session.Client().Database("ballots").Collection("polls")

	var p poll
	if err := respond.DecodeBody(r, &p); err != nil {
		respond.RespondErr(w, r, http.StatusBadRequest, "failed to read poll from request", err)
		return
	}
	apikey, ok := APIKey(r.Context())
	if ok {
		p.APIKey = apikey
	}
	p.ID = primitive.NewObjectID()
	result, err := c.InsertOne(r.Context(), p)
	if err != nil {
		respond.RespondErr(w, r, http.StatusInternalServerError, "filed to insert poll", err)
		return
	}
	log.Println("insert success to id, ", result.InsertedID)
	w.Header().Set("Location", "polls/"+p.ID.Hex())
	respond.Respond(w, r, http.StatusCreated, nil)
}

func (s *Server) handlePollsDelete(w http.ResponseWriter, r *http.Request) {
	session, err := s.db.StartSession()
	if err != nil {
		log.Fatalln(err)
	}
	defer session.EndSession(r.Context())
	c := session.Client().Database("ballots").Collection("polls")
	p := NewPath(r.URL.Path)
	if !p.HasID() {
		respond.RespondErr(w, r, http.StatusMethodNotAllowed, "Cannot delete all polls at the same time.")
		return
	}
	objectId, err := primitive.ObjectIDFromHex(p.Id)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = c.DeleteOne(r.Context(), bson.M{"_id": objectId})
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("delete success to id, ", objectId)
	respond.Respond(w, r, http.StatusOK, nil)
}
