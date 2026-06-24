package repository

type TicketRepo struct {
	repo *Repository
}

func NewTicketRepo(repo *Repository) *TicketRepo {
	return &TicketRepo{
		repo: repo,
	}
}
