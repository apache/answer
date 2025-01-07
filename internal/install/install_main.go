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
 */

package install

import (
	"fmt"
	"os"

	"github.com/apache/answer/internal/base/translator"
	"github.com/apache/answer/internal/cli"
)

var (
	port     = os.Getenv("INSTALL_PORT")
	confPath = ""
)

func Run(configPath string) {
	confPath = configPath
	// initialize translator for return internationalization error when installing.
	_, err := translator.NewTranslator(&translator.I18n{BundleDir: cli.I18nPath})
	if err != nil {
		panic(err)
	}

	// try to install by env
	if installByEnv, err := TryToInstallByEnv(); installByEnv && err != nil {
		fmt.Printf("[auto-install] try to init by env fail: %v\n", err)
	}

	installServer := NewInstallHTTPServer()
	if len(port) == 0 {
		port = "80"
	}
	fmt.Printf("[SUCCESS] answer installation service will run at: http://localhost:%s/install/ \n", port)
	if err = installServer.Run(":" + port); err != nil {
		panic(err)
	}
}
