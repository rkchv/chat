package err

type Code uint32

const (
	OK Code = iota
	// InvalidArgument Неверные параметры для операции/запроса
	InvalidArgument
	// NotFound Данные по запросу не найдены
	NotFound
	// DomainLogic Ошибка вызванная бизнес-логикой / инвариантом домена
	DomainLogic
	// AlreadyExists Запись / данные уже существуют
	AlreadyExists
	// PermissionDenied Недостаточно прав для совершения операции
	PermissionDenied
	// Unauthenticated Запрос не авторизован
	Unauthenticated
	// Canceled Контекст отменен
	Canceled
	// DeadlineExceeded Истекло время на запрос/операцию
	DeadlineExceeded
	// ResourceExhausted Превышено кол-во запросов / недостаточно ресурсов
	ResourceExhausted
	// Internal Внутренняя ошибка сервиса
	Internal
	// Unknown Неизвестная ошибка
	Unknown
)
