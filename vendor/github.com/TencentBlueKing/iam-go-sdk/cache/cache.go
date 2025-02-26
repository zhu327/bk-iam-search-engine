/*
 * TencentBlueKing is pleased to support the open source community by making 蓝鲸智云PaaS平台社区版 (BlueKing PaaS
 * Community Edition) available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package cache

import (
	"time"

	gocache "github.com/patrickmn/go-cache"
)

// Cache is the interface for common cache module
type Cache interface {
	Get(k string) (interface{}, bool)
	Set(k string, x interface{}, d time.Duration)
}

var c Cache

func init() {
	c = gocache.New(5*time.Minute, 10*time.Minute)
}

// Set set the key-value with ttl
func Set(k string, x interface{}, d time.Duration) {
	c.Set(k, x, d)
}

// Get get value of the key
func Get(k string) (interface{}, bool) {
	return c.Get(k)
}

// SetCache set the cache instance for the sdk
func SetCache(cache Cache) {
	c = cache
}
