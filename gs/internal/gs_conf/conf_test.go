/*
 * Copyright 2024 The Go-Spring Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gs_conf

import (
	"errors"
	"os"
	"testing"

	"github.com/go-spring/spring-core/conf"
	"github.com/lvan100/go-assert"
)

func clean() {
	os.Args = nil
	os.Clearenv()
	SysConf = conf.New()
}

func TestAppConfig(t *testing.T) {
	clean()

	t.Run("resolve error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "${a}")
		_, err := NewAppConfig().Refresh()
		assert.ThatError(t, err).Matches(`resolve string "\${a}" error << property a not exist`)
	})

	t.Run("resolve error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-REMOTE_DIR", "${a}")
		_, err := NewAppConfig().Refresh()
		assert.ThatError(t, err).Matches(`resolve string "\${a}" error << property a not exist`)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		_ = os.Setenv("GS_SPRING_APP_CONFIG-REMOTE_DIR", "./testdata/conf/remote")
		p, err := NewAppConfig().Refresh()
		assert.Nil(t, err)
		assert.That(t, p.Data()).Equal(map[string]string{
			"spring.app.config-local.dir":  "./testdata/conf",
			"spring.app.config-remote.dir": "./testdata/conf/remote",
			"spring.app.name":              "remote",
			"http.server.addr":             "0.0.0.0:8080",
		})
	})

	t.Run("merge error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_A", "a")
		_ = os.Setenv("GS_A_B", "a.b")
		_, err := NewAppConfig().Refresh()
		assert.ThatError(t, err).Matches("property conflict at path a.b")
	})

	t.Run("merge error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		_ = SysConf.Set("http.server[0].addr", "0.0.0.0:8080")
		_, err := NewAppConfig().Refresh()
		assert.ThatError(t, err).Matches("property conflict at path http.server.addr")
	})
}

func TestBootConfig(t *testing.T) {
	clean()

	t.Run("resolve error", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "${a}")
		_, err := NewBootConfig().Refresh()
		assert.ThatError(t, err).Matches(`resolve string "\${a}" error << property a not exist`)
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		p, err := NewBootConfig().Refresh()
		assert.Nil(t, err)
		assert.That(t, p.Data()).Equal(map[string]string{
			"spring.app.config-local.dir": "./testdata/conf",
			"spring.app.name":             "test",
			"http.server.addr":            "0.0.0.0:8080",
		})
	})

	t.Run("merge error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_A", "a")
		_ = os.Setenv("GS_A_B", "a.b")
		_, err := NewBootConfig().Refresh()
		assert.ThatError(t, err).Matches("property conflict at path a.b")
	})

	t.Run("merge error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		_ = os.Setenv("GS_SPRING_APP_CONFIG-LOCAL_DIR", "./testdata/conf")
		_ = SysConf.Set("http.server[0].addr", "0.0.0.0:8080")
		_, err := NewBootConfig().Refresh()
		assert.ThatError(t, err).Matches("property conflict at path http.server.addr")
	})
}

func TestPropertySources(t *testing.T) {
	clean()

	t.Run("add dir error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddDir("non_existent_dir")
		assert.That(t, 1).Equal(len(ps.extraDirs))
	})

	t.Run("add dir error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		assert.Panic(t, func() {
			ps.AddDir("./testdata/conf/app.properties")
		}, "should be a directory")
	})

	t.Run("add dir error - 3", func(t *testing.T) {
		t.Cleanup(clean)
		defer func() { osStat = os.Stat }()
		osStat = func(name string) (os.FileInfo, error) {
			return nil, errors.New("permission denied")
		}
		ps := NewPropertySources(ConfigTypeLocal, "app")
		assert.Panic(t, func() {
			ps.AddDir("./testdata/conf/app.properties")
		}, "permission denied")
	})

	t.Run("add file error - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("non_existent_file")
		assert.That(t, 1).Equal(len(ps.extraFiles))
	})

	t.Run("add file error - 2", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		assert.Panic(t, func() {
			ps.AddFile("./testdata/conf")
		}, "should be a file")
	})

	t.Run("add file error - 3", func(t *testing.T) {
		t.Cleanup(clean)
		defer func() { osStat = os.Stat }()
		osStat = func(name string) (os.FileInfo, error) {
			return nil, errors.New("permission denied")
		}
		ps := NewPropertySources(ConfigTypeLocal, "app")
		assert.Panic(t, func() {
			ps.AddFile("./testdata/conf")
		}, "permission denied")
	})

	t.Run("reset", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("./testdata/conf/app.properties")
		assert.That(t, 1).Equal(len(ps.extraFiles))
		ps.AddDir("./testdata/conf")
		assert.That(t, 1).Equal(len(ps.extraDirs))
		ps.Reset()
		assert.That(t, 0).Equal(len(ps.extraFiles))
		assert.That(t, 0).Equal(len(ps.extraDirs))
	})

	t.Run("getDefaultDir - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		dir, err := ps.getDefaultDir(conf.Map(nil))
		assert.Nil(t, err)
		assert.That(t, "./conf").Equal(dir)
	})

	t.Run("getDefaultDir - 2", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeRemote, "app")
		dir, err := ps.getDefaultDir(conf.Map(nil))
		assert.Nil(t, err)
		assert.That(t, "./conf/remote").Equal(dir)
	})

	t.Run("getFiles - 1", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		files, err := ps.getFiles("./conf", conf.Map(nil))
		assert.Nil(t, err)
		assert.That(t, files).Equal([]string{
			"./conf/app.properties",
			"./conf/app.yaml",
			"./conf/app.yml",
			"./conf/app.toml",
			"./conf/app.tml",
			"./conf/app.json",
		})
	})

	t.Run("getFiles - 2", func(t *testing.T) {
		t.Cleanup(clean)
		p := conf.Map(map[string]any{
			"spring.profiles.active": "dev,test",
		})
		ps := NewPropertySources(ConfigTypeLocal, "app")
		files, err := ps.getFiles("./conf", p)
		assert.Nil(t, err)
		assert.That(t, files).Equal([]string{
			"./conf/app.properties",
			"./conf/app.yaml",
			"./conf/app.yml",
			"./conf/app.toml",
			"./conf/app.tml",
			"./conf/app.json",
			"./conf/app-dev.properties",
			"./conf/app-dev.yaml",
			"./conf/app-dev.yml",
			"./conf/app-dev.toml",
			"./conf/app-dev.tml",
			"./conf/app-dev.json",
			"./conf/app-test.properties",
			"./conf/app-test.yaml",
			"./conf/app-test.yml",
			"./conf/app-test.toml",
			"./conf/app-test.tml",
			"./conf/app-test.json",
		})
	})

	t.Run("loadFiles", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("./testdata/conf/app.properties")
		files, err := ps.loadFiles(conf.Map(nil))
		assert.Nil(t, err)
		assert.That(t, 1).Equal(len(files))
	})

	t.Run("loadFiles - getDefaultDir error", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources("invalid", "app")
		_, err := ps.loadFiles(conf.Map(nil))
		assert.ThatError(t, err).Matches("unknown config type: invalid")
	})

	t.Run("loadFiles - getFiles error", func(t *testing.T) {
		t.Cleanup(clean)
		p := conf.Map(map[string]any{
			"spring.profiles.active": "${a}",
		})
		ps := NewPropertySources(ConfigTypeLocal, "app")
		_, err := ps.loadFiles(p)
		assert.ThatError(t, err).Matches(`resolve string "\${a}" error << property a not exist`)
	})

	t.Run("loadFiles - resolve error", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("./testdata/conf/app-${a}.properties")
		_, err := ps.loadFiles(conf.Map(nil))
		assert.ThatError(t, err).Matches("property a not exist")
	})

	t.Run("loadFiles - confLoad error", func(t *testing.T) {
		t.Cleanup(clean)
		ps := NewPropertySources(ConfigTypeLocal, "app")
		ps.AddFile("./testdata/conf/error.json")
		_, err := ps.loadFiles(conf.Map(nil))
		assert.ThatError(t, err).Matches("cannot unmarshal .*")
	})
}
