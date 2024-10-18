
<h1 >
<h1 align="center">
  <br>
  <a href="https://github.com/zaporter-work/viam-sds011"><img src="https://raw.githubusercontent.com/zaporter-work/viam-sds011/main/etc/sds011.jpg" alt="SDS011 image" width="200"></a>
  <br>
  SDS011 Air quality sensor module for Viam
  <br>
</h1>

This module implements the [`rdk:component:sensor` API](https://docs.viam.com/components/sensor) and provides two sensor models:

```
zaporter:sds011:v1
zaporter:sds011:v1-fake
```

The `zaporter:sds011:v1` model supports the SDS011 Nova PM air quality sensor.
The `zaporter:sds011:v1-fake` model can be used for testing the module without hardware.

> [!NOTE]
> For more information, see [Modular Resources](https://docs.viam.com/registry/#modular-resources).

## Configure your SDS011 sensor

Navigate to the **Config** tab of your machine's page in [the Viam app](https://app.viam.com/).
Click on the **Components** subtab and click **Create component**.
Select the `sensor` type, then select the `sds011:v1` model.
Click **Add module**, then enter a name for your sensor and click **Create**.

On the new component panel, copy and paste the following attribute template into your sensor’s **Attributes** box:

```json
{
  "usb_interface": "<PATH TO USB PORT WHERE YOUR SENSOR IS PLUGGED IN>"
}
```

### Attributes

The following attributes are available for `zaporter:sds011:v1` sensors:

| Name    | Type   | Inclusion    | Description |
| ------- | ------ | ------------ | ----------- |
| `usb_interface` | string | **Required** | Path to the USB port where your sensor is plugged in; see instructions below. |

To find the correct path, SSH to your board and run the following command:

```sh{class="command-line" data-prompt="$"}
ls /dev/serial/by-id
```

This should output a list of one or more USB devices attached to your board, for example `usb-1a86_USB_Serial-if00-port0`.
If the air quality sensor is the only device plugged into your board, you can be confident that the only device listed is the correct one.
If you have multiple devices plugged into different USB ports, you may need to do some trial and error or unplug something to figure out which path to use.

The `v1-fake` model also requires you to assign a value to the `usb_interface` attribute, but you can set it as any string since the fake model doesn't actually communicate with any real hardware.


### Example configuration

Example attribute configuration:

```json
{
  "usb_interface": "/dev/serial/by-id/usb-1a86_USB_Serial-if00-port0"
}
```

### Output

The sensor returns the following output:

```json5
{
  "pm_10": float64, 
  "pm_2_5": float64,
  "units": "μg/m³"
}
```

(for backwards compatability, this also returns `pm_2.5` but due to complexities with parsing that key, I renamed the key `pm_2_5`)

***I will likely remove the pm_2.5 key in the future. Do not use it. Use pm_2_5***

You can view sensor readings on [your machine's **Control** tab in the Viam app](https://app.viam.com/) or by using the [sensor API](https://docs.viam.com/components/sensor).
