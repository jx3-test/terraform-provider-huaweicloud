package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type DeleteVpcRequest struct {
	// VPC的资源ID

	VpcId string `json:"vpc_id"`
}

func (o DeleteVpcRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteVpcRequest struct{}"
	}

	return strings.Join([]string{"DeleteVpcRequest", string(data)}, " ")
}
