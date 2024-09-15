package hook

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	pubHook "github.com/alx99/ika/hook"
	"github.com/alx99/ika/internal/config"
	"github.com/alx99/ika/mocks"
)

func TestGetFactories(t *testing.T) {
	type args struct {
		ctx        context.Context
		hooks      map[string]pubHook.Factory
		namespaces config.Namespaces
	}
	tests := []struct {
		name    string
		args    args
		want    HookFactories
		wantErr bool
	}{
		{
			name: "Single namespace with one hook",
			args: args{
				ctx: context.Background(),
				hooks: map[string]pubHook.Factory{
					"hook1": mocks.NewFactoryMock(t),
				},
				namespaces: config.Namespaces{
					"namespace1": {
						Name: "namespace1",
						Hooks: config.Hooks{
							{
								Name:    "hook1",
								Enabled: config.NewNullable(true),
							},
						},
					},
				},
			},
			want: HookFactories{
				{
					Name:       "hook1",
					Namespaces: []string{"namespace1"},
					Factory:    mocks.NewFactoryMock(t),
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple namespaces with the same hook",
			args: args{
				ctx: context.Background(),
				hooks: map[string]pubHook.Factory{
					"hook1": mocks.NewFactoryMock(t),
				},
				namespaces: config.Namespaces{
					"namespace1": {
						Name: "namespace1",
						Hooks: config.Hooks{
							{
								Name:    "hook1",
								Enabled: config.NewNullable(true),
							},
						},
					},
					"namespace2": {
						Name: "namespace2",
						Hooks: config.Hooks{
							{
								Name:    "hook1",
								Enabled: config.NewNullable(true),
							},
						},
					},
				},
			},
			want: HookFactories{
				{
					Name:       "hook1",
					Namespaces: []string{"namespace1", "namespace2"},
					Factory:    mocks.NewFactoryMock(t),
				},
			},
			wantErr: false,
		},
		{
			name: "Multiple namespaces with different hooks",
			args: args{
				ctx: context.Background(),
				hooks: map[string]pubHook.Factory{
					"hook1": mocks.NewFactoryMock(t),
					"hook2": mocks.NewFactoryMock(t),
				},
				namespaces: config.Namespaces{
					"namespace1": {
						Name: "namespace1",
						Hooks: config.Hooks{
							{
								Name:    "hook1",
								Enabled: config.NewNullable(true),
							},
						},
					},
					"namespace2": {
						Name: "namespace2",
						Hooks: config.Hooks{
							{
								Name:    "hook2",
								Enabled: config.NewNullable(true),
							},
						},
					},
				},
			},
			want: HookFactories{
				{
					Name:       "hook1",
					Namespaces: []string{"namespace1"},
					Factory:    mocks.NewFactoryMock(t),
				},
				{
					Name:       "hook2",
					Namespaces: []string{"namespace2"},
					Factory:    mocks.NewFactoryMock(t),
				},
			},
			wantErr: false,
		},
		{
			name: "Hook not found",
			args: args{
				ctx: context.Background(),
				hooks: map[string]pubHook.Factory{
					"hook1": mocks.NewFactoryMock(t),
				},
				namespaces: config.Namespaces{
					"namespace1": {
						Name: "namespace1",
						Hooks: config.Hooks{
							{
								Name:    "hook2",
								Enabled: config.NewNullable(true),
							},
						},
					},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "No hooks enabled",
			args: args{
				ctx: context.Background(),
				hooks: map[string]pubHook.Factory{
					"hook1": mocks.NewFactoryMock(t),
				},
				namespaces: config.Namespaces{
					"namespace1": {
						Name:  "namespace1",
						Hooks: config.Hooks{},
					},
				},
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetFactories(tt.args.ctx, tt.args.hooks, tt.args.namespaces)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFactories() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFactories() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createHooks(t *testing.T) {
	type args struct {
		ctx       context.Context
		hooksCfg  config.Hooks
		factories HookFactories
	}
	tests := []struct {
		name            string
		args            args
		wantErr         bool
		runTearDown     bool
		wantTearDownErr bool
	}{
		{
			name: "Single hook setup successfully",
			args: args{
				ctx: context.Background(),
				hooksCfg: config.Hooks{
					{
						Name:    "hook1",
						Enabled: config.NewNullable(true),
					},
				},
				factories: HookFactories{
					{
						Name: "hook1",
						Factory: mocks.NewFactoryMock(t).
							NewMock.
							Expect(context.Background()).
							Return(
								mocks.NewHookMock(t).
									SetupMock.
									ExpectCtxParam1(context.Background()).
									Return(nil), nil),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Hook setup fails",
			args: args{
				ctx: context.Background(),
				hooksCfg: config.Hooks{
					{
						Name:    "hook1",
						Enabled: config.NewNullable(true),
					},
				},
				factories: HookFactories{
					{
						Name: "hook1",
						Factory: mocks.NewFactoryMock(t).NewMock.
							Expect(context.Background()).
							Return(
								mocks.NewHookMock(t).SetupMock.
									ExpectCtxParam1(context.Background()).
									Return(fmt.Errorf("setup error")), nil),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Hook creation fails",
			args: args{
				ctx: context.Background(),
				hooksCfg: config.Hooks{
					{
						Name:    "hook1",
						Enabled: config.NewNullable(true),
					},
				},
				factories: HookFactories{
					{
						Name: "hook1",
						Factory: mocks.NewFactoryMock(t).NewMock.
							Expect(context.Background()).
							Return(nil, fmt.Errorf("creation error")),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Multiple hooks setup successfully",
			args: args{
				ctx: context.Background(),
				hooksCfg: config.Hooks{
					{
						Name:    "hook1",
						Enabled: config.NewNullable(true),
					},
					{
						Name:    "hook2",
						Enabled: config.NewNullable(true),
					},
				},
				factories: HookFactories{
					{
						Name: "hook1",
						Factory: mocks.NewFactoryMock(t).NewMock.
							Expect(context.Background()).
							Return(
								mocks.NewHookMock(t).SetupMock.
									ExpectCtxParam1(context.Background()).
									Return(nil), nil),
					},
					{
						Name: "hook2",
						Factory: mocks.NewFactoryMock(t).NewMock.
							Expect(context.Background()).
							Return(
								mocks.NewHookMock(t).SetupMock.
									ExpectCtxParam1(context.Background()).
									Return(nil), nil),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "No hooks enabled",
			args: args{
				ctx:      context.Background(),
				hooksCfg: config.Hooks{},
				factories: HookFactories{
					{
						Name:    "hook1",
						Factory: mocks.NewFactoryMock(t),
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Teardown error",
			args: args{
				ctx: context.Background(),
				hooksCfg: config.Hooks{
					{
						Name:    "hook1",
						Enabled: config.NewNullable(true),
					},
				},
				factories: HookFactories{
					{
						Name: "hook1",
						Factory: mocks.NewFactoryMock(t).
							NewMock.
							Expect(context.Background()).
							Return(
								mocks.NewHookMock(t).
									SetupMock.
									ExpectCtxParam1(context.Background()).
									Return(nil).
									TeardownMock.
									ExpectCtxParam1(context.Background()).
									Return(fmt.Errorf("teardown error")), nil),
					},
				},
			},
			wantErr:         false,
			runTearDown:     true,
			wantTearDownErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, teardown, err := createHooks[pubHook.TransportHook](tt.args.ctx, tt.args.hooksCfg, tt.args.factories)
			if (err != nil) != tt.wantErr {
				t.Errorf("createHooks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.runTearDown {
				err := teardown(context.Background())
				if (err != nil) != tt.wantTearDownErr {
					t.Errorf("teardown() error = %v, wantErr %v", err, tt.wantTearDownErr)
					return
				}
			}
		})
	}
}
