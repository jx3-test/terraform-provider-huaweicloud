package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type ChangeFailoverStrategyRequest struct {
	// 语言

	XLanguage *ChangeFailoverStrategyRequestXLanguage `json:"X-Language,omitempty"`
	// 实例ID。

	InstanceId string `json:"instance_id"`

	Body *FailoverStrategyRequest `json:"body,omitempty"`
}

func (o ChangeFailoverStrategyRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "ChangeFailoverStrategyRequest struct{}"
	}

	return strings.Join([]string{"ChangeFailoverStrategyRequest", string(data)}, " ")
}

type ChangeFailoverStrategyRequestXLanguage struct {
	value string
}

type ChangeFailoverStrategyRequestXLanguageEnum struct {
	ZH_CN ChangeFailoverStrategyRequestXLanguage
	EN_US ChangeFailoverStrategyRequestXLanguage
}

func GetChangeFailoverStrategyRequestXLanguageEnum() ChangeFailoverStrategyRequestXLanguageEnum {
	return ChangeFailoverStrategyRequestXLanguageEnum{
		ZH_CN: ChangeFailoverStrategyRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: ChangeFailoverStrategyRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c ChangeFailoverStrategyRequestXLanguage) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *ChangeFailoverStrategyRequestXLanguage) UnmarshalJSON(b []byte) error {
	myConverter := converter.StringConverterFactory("string")
	if myConverter != nil {
		val, err := myConverter.CovertStringToInterface(strings.Trim(string(b[:]), "\""))
		if err == nil {
			c.value = val.(string)
			return nil
		}
		return err
	} else {
		return errors.New("convert enum data to string error")
	}
}
