package reports

// SharedPerson is what one person paid towards shared expenses.
type SharedPerson struct {
	PersonID     int64  `json:"person_id"`
	Name         string `json:"name"`
	AmountCents  int64  `json:"amount_cents"`  // includes pending estimates
	PendingCents int64  `json:"pending_cents"` // part of AmountCents
}

// SharedReport splits shared (non-personal) expenses between the people for
// dates From..To. Joint-paid expenses are not part of it.
type SharedReport struct {
	From         string         `json:"from"`
	To           string         `json:"to"`
	TotalCents   int64          `json:"total_cents"`
	PendingCents int64          `json:"pending_cents"`
	People       []SharedPerson `json:"people"`
}

// BuildShared totals people into the shared report for from..to.
func BuildShared(from, to string, people []SharedPerson) SharedReport {
	rep := SharedReport{From: from, To: to, People: people}
	if rep.People == nil {
		rep.People = []SharedPerson{}
	}
	for _, p := range rep.People {
		rep.TotalCents += p.AmountCents
		rep.PendingCents += p.PendingCents
	}
	return rep
}
