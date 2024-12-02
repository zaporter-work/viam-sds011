// Package sds011 is the package for sds011
package sds011

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/ryszard/sds011/go/sds011"

	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var (
	Model     = resource.NewModel("zaporter", "sds011", "v1")
	ModelFake = resource.NewModel("zaporter", "sds011", "v1-fake")
)

func init() {
	registration := resource.Registration[resource.Resource, *Config]{
		Constructor: func(ctx context.Context,
			deps resource.Dependencies,
			conf resource.Config,
			logger logging.Logger,
		) (resource.Resource, error) {
			return createComponent(ctx, deps, conf, logger, false)
		},
	}
	resource.RegisterComponent(sensor.API, Model, registration)

	registrationFake := resource.Registration[resource.Resource, *Config]{
		Constructor: func(ctx context.Context,
			deps resource.Dependencies,
			conf resource.Config,
			logger logging.Logger,
		) (resource.Resource, error) {
			return createComponent(ctx, deps, conf, logger, true)
		},
	}
	resource.RegisterComponent(sensor.API, ModelFake, registrationFake)
}

type component struct {
	resource.Named
	resource.AlwaysRebuild
	cfg          *Config
	isFake       bool
	sds011Sensor *sds011.Sensor

	cancelCtx  context.Context
	cancelFunc func()

	logger logging.Logger
}

func createComponent(_ context.Context,
	_ resource.Dependencies,
	conf resource.Config,
	logger logging.Logger,
	isFake bool,
) (sensor.Sensor, error) {
	newConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return nil, errors.Wrap(err, "create component failed due to config parsing")
	}

	var sensor *sds011.Sensor
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	instance := &component{
		Named:        conf.ResourceName().AsNamed(),
		cfg:          newConf,
		cancelCtx:    cancelCtx,
		cancelFunc:   cancelFunc,
		sds011Sensor: sensor,
		isFake:       isFake,
		logger:       logger,
	}
	if !isFake {
		if err := instance.setupSensor(); err != nil {
			if instance.sds011Sensor != nil {
				instance.sds011Sensor.Close()
			}
			return nil, err
		}
	}
	return instance, nil
}

func (c *component) Readings(ctx context.Context, extra map[string]interface{}) (valret map[string]interface{}, errret error) {
    defer func(){
        if r := recover(); r != nil {
            /*
12/2/2024, 10:40:44 AM error rdk.modmanager.zaporter_sds011.StdErr pexec/managed_process.go:277 \_ github.com/ryszard/sds011/go/sds011.(*Sensor).Get(0x4000a374c0?)

12/2/2024, 10:40:44 AM error rdk.modmanager.zaporter_sds011.StdErr pexec/managed_process.go:277 \_ /home/zack/go/pkg/mod/github.com/ryszard/sds011@v0.0.0-20170226135337-5d7058e01434/go/sds011/sds011.go:72 +0x80

12/2/2024, 10:40:44 AM error rdk.modmanager.zaporter_sds011.StdErr pexec/managed_process.go:277 \_ github.com/ryszard/sds011/go/sds011.(*response).PM25(0x4000a374c0?)

12/2/2024, 10:40:44 AM error rdk.modmanager.zaporter_sds011.StdErr pexec/managed_process.go:277 \_ goroutine 35040 [running]:

12/2/2024, 10:40:44 AM error rdk.modmanager.zaporter_sds011.StdErr pexec/managed_process.go:277 \_

12/2/2024, 10:40:44 AM error rdk.modmanager.zaporter_sds011.StdErr pexec/managed_process.go:277 \_ panic: access to field that doesn't work with this type of response &sds011.response{Header:0xaa, Command:0xc5, Data:[6]uint8{0x2, 0x1, 0x1, 0x0, 0x77, 0x95}, CheckSum:0x10, Tail:0xab} 
            */
            errret= errors.New(fmt.Sprintf("Panic when calling Readings() %+v", r))
        }
    }()
	if c.isFake {
		return map[string]interface{}{
			"pm_10": 10.0,
			"units": "μg/m³",
		}, nil
	}
	reading, err := c.sds011Sensor.Query()
	if err != nil {
		// try resetting the sensor
		if err2 := c.setupSensor(); err2 != nil {
			return nil, errors.Wrap(err, err2.Error())
		}
		reading, err = c.sds011Sensor.Query()
		if err != nil {
			return nil, err
		}
	}
	return map[string]interface{}{
		"pm_10":  reading.PM10,
		"pm_2.5": reading.PM25,
		"pm_2_5": reading.PM25,
		"units":  "μg/m³",
	}, nil
}

func (c *component) setupSensor() error {
	c.logger.Info("setting up sensor\n")
	if c.sds011Sensor != nil {
		c.sds011Sensor.Close()
	}
	var err error
	c.sds011Sensor, err = sds011.New(c.cfg.USBInterface)
	if err != nil {
		return errors.Wrapf(err, "unable to connect to interface %q", c.cfg.USBInterface)
	}
	if val, err := c.sds011Sensor.IsAwake(); err != nil || !val {
		if err != nil {
			return errors.Wrap(err, "reading sensor awakeness")
		}
		if err := c.sds011Sensor.Awake(); err != nil {
			return errors.Wrap(err, "unable to set the sensor to awake")
		}
	}
	if err := c.sds011Sensor.MakePassive(); err != nil {
		return errors.Wrap(err, "unable to set the sensor to passive")
	}
	return nil
}

// DoCommand sends/receives arbitrary data.
func (c *component) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	return make(map[string]interface{}), nil
}

// Close must safely shut down the resource and prevent further use.
// Close must be idempotent.
// Later reconfiguration may allow a resource to be "open" again.
func (c *component) Close(ctx context.Context) error {
	c.cancelFunc()
	if c.sds011Sensor != nil {
		c.sds011Sensor.Close()
	}
	c.logger.Info("closing\n")
	return nil
}
