package domain

import "strings"

type RolePermissions struct {
	RolePermissions map[string][]string
}

func (p RolePermissions) IsAuthorisedFor(role string, routeName string) bool {
	perms := p.RolePermissions[role]
	for _, r := range perms {
		if r == strings.TrimSpace(routeName) {
			return true
		}
	}
	return false
}

func GetRolePermissions() RolePermissions {
	return RolePermissions{map[string][]string{
		"admin": {"GetAllCustomers", "GetCustomer", "NewAccount", "NewTransaction"},
		"user":  {"GetCustomer", "NewTransaction"},
	}}
}
