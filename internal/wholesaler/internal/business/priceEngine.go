package business

import (
	"errors"
	"gomarketplace_api/internal/wholesaler/internal/models"
	"gomarketplace_api/internal/wholesaler/internal/storage/repositories"
	"log"
	"math"
	"math/rand"
	"time"
)

type PriceEngine struct {
	repo *repositories.PriceRepository
}

func NewPriceEngine(repo *repositories.PriceRepository) *PriceEngine {
	return &PriceEngine{repo}
}

func (e *PriceEngine) GetPrices(all bool) (interface{}, error) {
	priceData, err := e.repo.GetPrices()
	prices := make(map[int]PriceResult)

	if err != nil {
		return nil, err
	}

	for k, v := range priceData {
		if !all && v <= 0 {
			continue
		}
		prices[k] = e.CalculatePrices(v)
	}

	return prices, nil
}

func (e *PriceEngine) GetPriceById(id int) (float32, error) {
	return e.GetPriceById(id)
}

type PriceResult struct {
	X int `json:"X"` // это премиум цена на Озоне
	Y int `json:"Y"` // это завышенная цена (до скидки)
	Z int `json:"Z"` // это финальная цена
	T int `json:"T"` // размер 5% премиум скидки в руб
	Q int `json:"Q"` // желаемая чистая прибыль в руб
	S int `json:"S"` // расчёт стоимости доставки
	R int `json:"R"` // цена товара без учёта доставки
}

var (
	D = []int{38, 37, 34, 29, 26, 25, 25,
		25, 24, 22, 22, 21, 21, 21,
		21, 20, 20, 19, 18, 17, 16,
		15, 14, 13}
	MIN_Q                    = 30.0
	MIN_S                    = 55.0
	MAX_S                    = 205.0
	MAX_U_COEFFICIENT        = 0.5
	Q_COEFFICIENT            = 1.01
	T_MIN                    = 20.0
	T_MAX                    = 500.0
	DIVISION_ROUNDING        = 100.0
	ADDING_PRICE_COEFFICIENT = 1.05
)

func getDValue(P float64) (int, error) {
	switch {
	case P >= 1 && P <= 100:
		return D[0], nil
	case P >= 101 && P <= 200:
		return D[1], nil
	case P >= 201 && P <= 400:
		return D[2], nil
	case P >= 401 && P <= 700:
		return D[3], nil
	case P >= 701 && P <= 1000:
		return D[4], nil
	case P >= 1001 && P <= 1300:
		return D[5], nil
	case P >= 1301 && P <= 1600:
		return D[6], nil
	case P >= 1601 && P <= 1900:
		return D[7], nil
	case P >= 1901 && P <= 2300:
		return D[8], nil
	case P >= 2301 && P <= 2700:
		return D[9], nil
	case P >= 2701 && P <= 3000:
		return D[10], nil
	case P >= 3001 && P <= 4000:
		return D[11], nil
	case P >= 4001 && P <= 5000:
		return D[12], nil
	case P >= 5001 && P <= 6000:
		return D[13], nil
	case P >= 6001 && P <= 7000:
		return D[14], nil
	case P >= 7001 && P <= 8000:
		return D[15], nil
	case P >= 8001 && P <= 9000:
		return D[16], nil
	case P >= 9001 && P <= 10000:
		return D[17], nil
	case P >= 10001 && P <= 12000:
		return D[18], nil
	case P >= 12001 && P <= 14000:
		return D[19], nil
	case P >= 14001 && P <= 16000:
		return D[20], nil
	case P >= 16001 && P <= 18000:
		return D[21], nil
	case P >= 18001 && P <= 20000:
		return D[22], nil
	case P >= 20001:
		return D[23], nil
	default:
		return 0, errors.New("purchase price out of range")
	}
}

func calculateQ(P float64) float64 {
	D, _ := getDValue(P)
	Q := math.Ceil(P*float64(D)/100 + rand.Float64()*Q_COEFFICIENT)
	if Q < MIN_Q {
		return MIN_Q
	}
	return Q
}

func calculateR(P, Q float64) float64 {
	return math.Round((P+Q+55+20)*100/(100-23)*DIVISION_ROUNDING) / DIVISION_ROUNDING
}

func calculateS(R float64) float64 {
	S := R / 100 * 5
	if S < MIN_S {
		return MIN_S
	} else if S > MAX_S {
		return MAX_S
	}
	return math.Round(S*DIVISION_ROUNDING) / DIVISION_ROUNDING
}

func calculateZ(P, Q, S float64) float64 {
	return math.Round((P+Q+S+25+20)*100/(100-23)*DIVISION_ROUNDING*ADDING_PRICE_COEFFICIENT) / DIVISION_ROUNDING
}

func calculateX(Z, T, Q float64) float64 {
	if T >= MAX_U_COEFFICIENT*Q {
		return 0
	}
	return Z - T - 1
}

func calculateT(Z float64) float64 {
	T := Z * 5 / 100
	if T >= T_MAX {
		return T_MAX
	} else if T <= T_MIN {
		return T_MIN
	}
	return math.Round(T*DIVISION_ROUNDING) / DIVISION_ROUNDING
}

func (e *PriceEngine) CalculatePrices(P int) PriceResult {
	p := float64(P)
	Q := calculateQ(p)
	R := calculateR(p, Q)
	S := calculateS(R)
	Z := calculateZ(p, Q, S)
	T := calculateT(Z)
	X := calculateX(Z, T, Q)
	Y := math.Round(Z * DIVISION_ROUNDING / DIVISION_ROUNDING)
	return PriceResult{
		X: int(X), Y: int(Y), Z: int(Z), T: int(T), Q: int(Q), S: int(S), R: int(R),
	}
}

func (e *PriceEngine) GetProductPriceByID(id int) (*models.Price, error) {
	if id <= 0 {
		return nil, errors.New("invalid product ID")
	}

	price, err := e.repo.GetPriceByProductID(id)
	if err != nil {
		return nil, err
	}

	if price == nil {
		return nil, errors.New("product not found")
	}

	log.Printf("Retrieved price with product with ID: %d", id)
	return price, nil
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
