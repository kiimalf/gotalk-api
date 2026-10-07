package service

import (
	"strings"

	"gotalk-api/app/model"
	"gotalk-api/helper"
)

func CanAccessUser(
	current model.AuthUser,
	targetID int,
	perms *helper.PermissionSet,
	anyPermission string,
) bool {
	if current.UserID == targetID {
		return true
	}
	return perms.Can(current.Role, anyPermission)
}

func ValidateAssignRole(
	current model.AuthUser,
	targetID int,
	req model.AssignRoleRequest,
	perms *helper.PermissionSet,
) map[string]string {
	errs := map[string]string{}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		errs["role"] = "Wajib diisi"
		return errs
	}

	if !perms.IsKnownRole(role) {
		errs["role"] = "Role tidak dikenal, pilih salah satu dari: " +
			strings.Join(perms.KnownRoles(), ", ")
	}

	if current.UserID == targetID {
		errs["role"] = "Tidak boleh mengubah role diri sendiri"
	}

	return errs
}
