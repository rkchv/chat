package rate_limiter

import (
	"sync/atomic"
	"time"
)

// RateLimiter оперирует входящим потоком запросов на основе корзины с токенами
type RateLimiter struct {
	// maxTokens максимальное число запросов в заданный период
	maxTokens uint32
	// period интервал времени для которого работает ограничение
	// например для 1 RPS period будет time.Second
	period time.Duration
	// bucketTokens корзина с токенами, в нее идет пополнение не более чем maxTokens
	bucketTokens atomic.Uint32
	// lastFill когда было последнее заполнение (генерация) корзины токенами
	lastFill time.Time
}

// New Создает новый экземпляр RateLimiter
func New(maxTokens uint, period time.Duration) *RateLimiter {
	rl := &RateLimiter{
		lastFill:  time.Now(),
		period:    period,
		maxTokens: uint32(maxTokens),
	}

	rl.bucketTokens.Store(uint32(maxTokens))

	return rl
}

// Allow проверяет есть ли в корзине еще токены (не превышена ли скорость запросов)
// так же непосредственно при вызове производит дозаполнение корзины
func (r *RateLimiter) Allow() bool {
	// Считаем сколько токенов должны были сгенерировать за время прошедшее с последнего заполнения корзины(генерации)
	sinceRefill := uint32(time.Since(r.lastFill) / r.period) //сколько прошло заданных "периодов"
	tokensToAdd := r.maxTokens * sinceRefill                 //кол-во токенов которое должно получится за эти периоды
	currentTokens := r.bucketTokens.Load() + tokensToAdd     //общий остаток токенов
	// Если недостаточно, запросы пришли чаще
	if currentTokens < 1 {
		return false
	}

	// Если токены забиты под завязку, время заполнения = текущее время
	if currentTokens > r.maxTokens {
		r.lastFill = time.Now()
		r.bucketTokens.Store(r.maxTokens - 1)
	} else {
		//Если место в корзине есть, то кладем столько токенов, сколько сгенерировались за прошедший период
		//один вычитаем для текущего запроса
		//время последнего заполнения пишем не как time.Now()
		//а как отношение кол-ва сгенер. токенов на заданный интервал
		//мы как бы имитируем тем самым ticker
		deltaRefills := tokensToAdd / r.maxTokens
		deltaTime := time.Duration(deltaRefills) * r.period
		r.lastFill = r.lastFill.Add(deltaTime)
		r.bucketTokens.Store(currentTokens - 1)
	}

	return true
}
