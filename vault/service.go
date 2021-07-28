package vault

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	Hash(ctx context.Context, password string) (string, error)
	Validate(ctx context.Context, password, hash string) (bool, error)
}

type vaultService struct{}

func NewService() Service {
	return vaultService{}
}

// application에 암호 저장하는 방법 확인 --> https://codahale.com/how-to-safety-store-a-password/
func (vaultService) Hash(ctx context.Context, password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), err
}

func (vaultService) Validate(ctx context.Context, password, hash string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, nil
}

type hashRequest struct {
	Password string `json:"password"`
}
type hashResponse struct {
	Hash string `json:"hash"`
	Err  string `json:"err,omitempty`
}

func decodeHashRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req hashRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

type validateRequest struct {
	Password string `json:"password"`
	Hash     string `json:"hash"`
}
type validateResponse struct {
	Valid bool   `json:"valid"`
	Err   string `json:"err,omitempty"`
}

func decodeValidateRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	var req validateRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

//서비스 메소드용 엔드포인트
// 항상 err -> nil, hashResponse 객체에 err의 string을 포함, 여기서 err이 nil이 아니면 미들웨어에서 오류 --> 좀더 고민해보자..
func MakeHashEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(hashRequest)
		v, err := srv.Hash(ctx, req.Password)
		if err != nil {
			return hashResponse{v, err.Error()}, nil
		}
		return hashResponse{v, ""}, nil
	}
}

func MakeValidateEndpoint(srv Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(validateRequest)
		ok, err := srv.Validate(ctx, req.Password, req.Hash)
		if err != nil {
			return validateResponse{false, err.Error()}, nil
		}
		return validateResponse{ok, ""}, nil
	}
}

// endpoint를 서비스 구현으로 맵핑
type Endpoints struct {
	HashEndPoint     endpoint.Endpoint
	ValidateEndPoint endpoint.Endpoint
}

func (e Endpoints) Hash(ctx context.Context, password string) (string, error) {
	req := hashRequest{Password: password}
	resp, err := e.HashEndPoint(ctx, req)
	if err != nil {
		return "", err
	}
	hashResp := resp.(hashResponse)
	if hashResp.Err != "" {
		return "", errors.New(hashResp.Err)
	}
	return hashResp.Hash, nil
}

func (e Endpoints) Validate(ctx context.Context, password, hash string) (bool, error) {
	req := validateRequest{Password: password, Hash: hash}
	resp, err := e.ValidateEndPoint(ctx, req)
	if err != nil {
		return false, err
	}
	validateResp := resp.(validateResponse)
	if validateResp.Err != "" {
		return false, errors.New(validateResp.Err)
	}
	return validateResp.Valid, nil
}
