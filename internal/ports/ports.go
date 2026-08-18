// Package ports defines the interfaces the service layer depends on. They
// live here (and not inside the concrete implementation packages) so that
// domain services can take them as parameters without importing the concrete
// types — and without creating import cycles with the service container.
package ports

import (
	"time"

	"github.com/jpporta/ticket-control/internal/printer"
)

// Printer is the port the service layer depends on. The concrete
// *printer.Printer satisfies it; tests can supply stubs.
type Printer interface {
	PrintTask(id int32, title, description string, priority int32, createdBy string, createdAt time.Time) error
	PrintTasks(tasks []printer.TaskInput) error
	PrintList(list printer.ListInput) error
	PrintLink(link printer.LinkInput) error
	PrintEndOfDay(input printer.EndOfDayInput) error
	PrintLetter(input printer.LetterInput) error
	PrintBip() error
	Toggle(enabled bool)
}

// Compile-time assertion that *printer.Printer satisfies the Printer port.
var _ Printer = (*printer.Printer)(nil)
