package orm

type Engine interface {
	CreateTable()
	DropTable()
	AddColumn()
}
