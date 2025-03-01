package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/converter"

	"strings"
)

// Request Object
type StartResizeFlavorActionRequest struct {
	// 语言

	XLanguage *StartResizeFlavorActionRequestXLanguage `json:"X-Language,omitempty"`
	// 实例ID。

	InstanceId string `json:"instance_id"`

	Body *ResizeFlavorRequest `json:"body,omitempty"`
}

func (o StartResizeFlavorActionRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "StartResizeFlavorActionRequest struct{}"
	}

	return strings.Join([]string{"StartResizeFlavorActionRequest", string(data)}, " ")
}

type StartResizeFlavorActionRequestXLanguage struct {
	value string
}

type StartResizeFlavorActionRequestXLanguageEnum struct {
	ZH_CN StartResizeFlavorActionRequestXLanguage
	EN_US StartResizeFlavorActionRequestXLanguage
}

func GetStartResizeFlavorActionRequestXLanguageEnum() StartResizeFlavorActionRequestXLanguageEnum {
	return StartResizeFlavorActionRequestXLanguageEnum{
		ZH_CN: StartResizeFlavorActionRequestXLanguage{
			value: "zh-cn",
		},
		EN_US: StartResizeFlavorActionRequestXLanguage{
			value: "en-us",
		},
	}
}

func (c StartResizeFlavorActionRequestXLanguage) MarshalJSON() ([]byte, error) {
	return utils.Marshal(c.value)
}

func (c *StartResizeFlavorActionRequestXLanguage) UnmarshalJSON(b []byte) error {
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
