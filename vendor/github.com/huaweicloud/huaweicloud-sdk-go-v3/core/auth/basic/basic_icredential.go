// Copyright 2020 Huawei Technologies Co.,Ltd.
//
// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package basic

import (
	"fmt"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/cache"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/internal"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/signer"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/impl"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/request"
	"strings"
	"time"
)

const (
	ProjectIdInHeader     = "X-Project-Id"
	SecurityTokenInHeader = "X-Security-Token"
	ContentTypeInHeader   = "Content-Type"
)

var DefaultDerivedPredicate = auth.GetDefaultDerivedPredicate()

type Credentials struct {
	IamEndpoint      string
	AK               string
	SK               string
	ProjectId        string
	SecurityToken    string
	DerivedPredicate func(*request.DefaultHttpRequest) bool

	derivedAuthServiceName string
	regionId               string
	expiredAt              int64
}

func (s Credentials) ProcessAuthParams(client *impl.DefaultHttpClient, region string) auth.ICredential {
	if s.ProjectId != "" {
		return s
	}

	authCache := cache.GetCache()
	akWithName := s.AK + region
	if projectId, ok := authCache.GetAuth(akWithName); ok {
		s.ProjectId = projectId
		return s
	}

	derivedPredicate := s.DerivedPredicate
	s.DerivedPredicate = nil

	req, err := s.ProcessAuthRequest(client, internal.GetKeystoneListProjectsRequest(s.IamEndpoint, region))
	if err != nil {
		panic(fmt.Sprintf("failed to get project id, %s", err.Error()))
	}

	id, err := internal.KeystoneListProjects(client, req)
	if err != nil {
		panic(fmt.Sprintf("failed to get project id, %s", err.Error()))
	}

	s.ProjectId = id
	authCache.PutAuth(akWithName, id)

	s.DerivedPredicate = derivedPredicate

	return s
}

func (s Credentials) ProcessAuthRequest(client *impl.DefaultHttpClient, req *request.DefaultHttpRequest) (*request.DefaultHttpRequest, error) {
	reqBuilder := req.Builder()

	if s.NeedUpdate() {
		err := s.UpdateCredential(client)
		if err != nil {
			return nil, err
		}
	}

	if s.ProjectId != "" {
		reqBuilder = reqBuilder.
			AddAutoFilledPathParam("project_id", s.ProjectId).
			AddHeaderParam(ProjectIdInHeader, s.ProjectId)
	}

	if s.SecurityToken != "" {
		reqBuilder.AddHeaderParam(SecurityTokenInHeader, s.SecurityToken)
	}

	if _, ok := req.GetHeaderParams()[ContentTypeInHeader]; ok {
		if !strings.Contains(req.GetHeaderParams()[ContentTypeInHeader], "application/json") {
			reqBuilder.AddHeaderParam("X-Sdk-Content-Sha256", "UNSIGNED-PAYLOAD")
		}
	}

	var (
		headerParams map[string]string
		err          error
	)

	if s.IsDerivedAuth(req) {
		headerParams, err = signer.SignDerived(reqBuilder.Build(), s.AK, s.SK, s.derivedAuthServiceName, s.regionId)
	} else {
		headerParams, err = signer.Sign(reqBuilder.Build(), s.AK, s.SK)
	}

	if err != nil {
		return nil, err
	}

	for key, value := range headerParams {
		req.AddHeaderParam(key, value)
	}
	return req, nil
}

func (s Credentials) ProcessDerivedAuthParams(derivedAuthServiceName, regionId string) auth.ICredential {
	if s.derivedAuthServiceName == "" {
		s.derivedAuthServiceName = derivedAuthServiceName
	}

	if s.regionId == "" {
		s.regionId = regionId
	}

	return s
}

func (s Credentials) IsDerivedAuth(httpRequest *request.DefaultHttpRequest) bool {
	if s.DerivedPredicate == nil {
		return false
	}

	return s.DerivedPredicate(httpRequest)
}

func (s Credentials) NeedUpdate() bool {
	if s.AK == "" || s.SK == "" {
		return true
	}

	if s.expiredAt == 0 {
		return false
	}

	return s.expiredAt-time.Now().Unix() < 60
}

func (s *Credentials) UpdateCredential(client *impl.DefaultHttpClient) error {
	credential, err := internal.GetTemporaryCredential(client)
	if err != nil {
		return err
	}

	s.AK = credential.Access
	s.SK = credential.Secret
	s.SecurityToken = credential.Securitytoken
	location, err := time.ParseInLocation(`2006-01-02T15:04:05Z`, credential.ExpiresAt, time.UTC)
	if err != nil {
		return err
	}
	s.expiredAt = location.Unix()

	return nil
}

type CredentialsBuilder struct {
	Credentials Credentials
}

func NewCredentialsBuilder() *CredentialsBuilder {
	return &CredentialsBuilder{Credentials: Credentials{
		IamEndpoint: internal.DefaultIamEndpoint,
	}}
}

func (builder *CredentialsBuilder) WithIamEndpointOverride(endpoint string) *CredentialsBuilder {
	builder.Credentials.IamEndpoint = endpoint
	return builder
}

func (builder *CredentialsBuilder) WithAk(ak string) *CredentialsBuilder {
	builder.Credentials.AK = ak
	return builder
}

func (builder *CredentialsBuilder) WithSk(sk string) *CredentialsBuilder {
	builder.Credentials.SK = sk
	return builder
}

func (builder *CredentialsBuilder) WithProjectId(projectId string) *CredentialsBuilder {
	builder.Credentials.ProjectId = projectId
	return builder
}

func (builder *CredentialsBuilder) WithSecurityToken(token string) *CredentialsBuilder {
	builder.Credentials.SecurityToken = token
	return builder
}

func (builder *CredentialsBuilder) WithDerivedPredicate(derivedPredicate func(*request.DefaultHttpRequest) bool) *CredentialsBuilder {
	builder.Credentials.DerivedPredicate = derivedPredicate
	return builder
}

func (builder *CredentialsBuilder) Build() Credentials {
	return builder.Credentials
}
