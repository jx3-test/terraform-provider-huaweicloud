package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

//
type PolicyRoleResult struct {
	// 自定义策略所属账号ID。

	DomainId string `json:"domain_id"`
	// 自定义策略的引用次数。

	References *int32 `json:"references,omitempty"`
	// 自定义策略更新时间。

	UpdatedTime *string `json:"updated_time,omitempty"`
	// 自定义策略创建时间。

	CreatedTime *string `json:"created_time,omitempty"`
	// 自定义策略的中文描述信息。

	DescriptionCn *string `json:"description_cn,omitempty"`
	// 自定义策略所在目录。

	Catalog string `json:"catalog"`
	// 自定义策略名。

	Name string `json:"name"`
	// 自定义策略的描述信息。

	Description string `json:"description"`

	Links *LinksSelf `json:"links"`
	// 自定义策略ID。

	Id string `json:"id"`
	// 自定义策略展示名。

	DisplayName string `json:"display_name"`
	// 自定义策略的显示模式。 > - AX表示在domain层显示。 > - XA表示在project层显示。 > - 自定义策略的显示模式只能为AX或者XA，不能在domain层和project层都显示（AA），或者在domain层和project层都不显示（XX）。

	Type string `json:"type"`

	Policy *ServicePolicy `json:"policy"`
}

func (o PolicyRoleResult) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "PolicyRoleResult struct{}"
	}

	return strings.Join([]string{"PolicyRoleResult", string(data)}, " ")
}
