/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package policy

import (
	"context"

	apisecurity "github.com/polarismesh/specification/source/go/api/v1/security"

	"github.com/polarismesh/polaris/auth"
	cachetypes "github.com/polarismesh/polaris/cache/api"
	authcommon "github.com/polarismesh/polaris/common/model/auth"
	"github.com/polarismesh/polaris/common/utils"
	"github.com/polarismesh/polaris/store"
)

type DefaultPolicyHelper struct {
	options  *AuthConfig
	storage  store.Store
	cacheMgr cachetypes.CacheManager
	checker  auth.AuthChecker
}

func (h *DefaultPolicyHelper) GetRole(id string) *authcommon.Role {
	return h.cacheMgr.Role().GetRole(id)
}

func (h *DefaultPolicyHelper) GetPolicyRule(id string) *authcommon.StrategyDetail {
	return h.cacheMgr.AuthStrategy().GetPolicyRule(id)
}

// CreatePrincipal 创建 principal 的默认 policy 资源
func (h *DefaultPolicyHelper) CreatePrincipal(ctx context.Context, tx store.Tx, p authcommon.Principal) error {
	if err := h.storage.AddStrategy(tx, defaultPrincipalPolicy(p)); err != nil {
		return err
	}
	return nil
}

func defaultPrincipalPolicy(p authcommon.Principal) *authcommon.StrategyDetail {
	// Create the user's default weight policy
	ruleId := utils.NewUUID()

	resources := []authcommon.StrategyResource{}
	if p.PrincipalType == authcommon.PrincipalUser {
		resources = append(resources, authcommon.StrategyResource{
			StrategyID: ruleId,
			ResType:    int32(apisecurity.ResourceType_Users),
			ResID:      p.PrincipalID,
		})
	}

	return &authcommon.StrategyDetail{
		ID:         ruleId,
		Name:       authcommon.BuildDefaultStrategyName(authcommon.PrincipalUser, p.Name),
		Action:     apisecurity.AuthAction_ALLOW.String(),
		Default:    true,
		Owner:      p.Owner,
		Revision:   utils.NewUUID(),
		Source:     "Polaris",
		Resources:  resources,
		Principals: []authcommon.Principal{p},
		CalleeMethods: []string{
			// 用户操作权限
			string(authcommon.DescribeUsers),
			string(authcommon.DescribeUserToken),
			string(authcommon.UpdateUser),
			string(authcommon.UpdateUserPassword),
			string(authcommon.EnableUserToken),
			string(authcommon.ResetUserToken),
			// 鉴权策略
			string(authcommon.DescribeAuthPolicies),
			string(authcommon.DescribeAuthPolicyDetail),
			// 角色
			string(authcommon.DescribeAuthRoles),
		},
		Valid:   true,
		Comment: "default principal auth policy rule",
	}
}

// CleanPrincipal 清理 principal 所关联的 policy、role 资源
func (h *DefaultPolicyHelper) CleanPrincipal(ctx context.Context, tx store.Tx, p authcommon.Principal) error {
	if err := h.storage.CleanPrincipalPolicies(tx, p); err != nil {
		return err
	}

	if err := h.storage.CleanPrincipalRoles(tx, &p); err != nil {
		return err
	}
	return nil
}
