package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type RemoveAllProjectsPermissionFromAgencyRequest struct {
	// 委托ID，获取方式请参见：[获取委托名、委托ID](https://support.huaweicloud.com/api-iam/iam_17_0002.html)。

	AgencyId string `json:"agency_id"`
	// 账号ID，获取方式请参见：[获取账号ID](https://support.huaweicloud.com/api-iam/iam_17_0002.html)。

	DomainId string `json:"domain_id"`
	// 权限ID，获取方式请参见：[获取权限名、权限ID](https://support.huaweicloud.com/api-iam/iam_10_0001.html)。

	RoleId string `json:"role_id"`
}

func (o RemoveAllProjectsPermissionFromAgencyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "RemoveAllProjectsPermissionFromAgencyRequest struct{}"
	}

	return strings.Join([]string{"RemoveAllProjectsPermissionFromAgencyRequest", string(data)}, " ")
}
