package database

type Assembly struct {
	Name                string
	Assembly_id         int
	CLR_name            string
	Permission_set_desc string
	Is_user_defined     bool
}
