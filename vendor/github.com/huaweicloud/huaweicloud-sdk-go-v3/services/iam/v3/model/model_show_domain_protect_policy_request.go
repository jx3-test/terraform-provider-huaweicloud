package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ShowDomainProtectPolicyRequest struct {
	// 待查询的账号ID，获取方式请参见：[获取账号、IAM用户、项目、用户组、委托的名称和ID](https://support.huaweicloud.com/api-iam/iam_17_0002.html)。

	DomainId string `json:"domain_id"`
}

func (o ShowDomainProtectPolicyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ShowDomainProtectPolicyRequest struct{}"
	}

	return strings.Join([]string{"ShowDomainProtectPolicyRequest", string(data)}, " ")
}
