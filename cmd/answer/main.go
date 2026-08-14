/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
	根据一个或多个贡献者许可协议，
	本文件已授权给 Apache 软件基金会（ASF）。
	有关版权归属的更多信息，
	请参阅随本作品一起分发的 NOTICE 文件。
	Apache 软件基金会依据 Apache License 2.0 版本
	（以下简称“许可证”）向你授权使用本文件。
	除非符合该许可证的规定，
	否则你不得使用本文件。
	你可以在以下地址获取许可证副本：
	http://www.apache.org/licenses/LICENSE-2.0
	除非适用法律要求，或双方另有书面约定，
	根据本许可证分发的软件均按“原样（AS IS）”提供，
	不提供任何形式的明示或默示担保或条件。
	有关许可证所规定的具体权限和限制，
	请参阅许可证正文。
 */

//go:generate go run github.com/swaggo/swag/cmd/swag init -g ./cmd/answer/main.go -d ../../ -o ../../docs

package main

import (
	answercmd "github.com/apache/answer/cmd"
)

// main godoc
// @title Apache Answer
// @description Apache Answer API
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	answercmd.Main()
}
