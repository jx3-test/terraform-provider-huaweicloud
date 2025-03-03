package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type ListInstantQueryAomPromGetRequest struct {
	// PromQL表达式(参考https://prometheus.io/docs/prometheus/latest/querying/basics/)。

	Query string `json:"query"`
	// 指定用于计算 PromQL 的时间戳，(Unix时间戳格式，单位：秒）。

	Time *string `json:"time,omitempty"`
}

func (o ListInstantQueryAomPromGetRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ListInstantQueryAomPromGetRequest struct{}"
	}

	return strings.Join([]string{"ListInstantQueryAomPromGetRequest", string(data)}, " ")
}
