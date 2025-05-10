package errors

type OrganizationNotFoundError struct{}

func (m *OrganizationNotFoundError) Error() string {
	return "Organization not found"
}
