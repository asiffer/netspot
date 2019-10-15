## Routes


<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #29BEB0; margin-bottom: 0.5em; border-radius: 8px;">GET /api/config</div>

*Return the current configuration*

**Responses**

| Code | Description | Content | Return |
|------|-------------|---------|--------|
| **200** | The configuration of netspot | `application/json` | [Config](#config) |
| **405** | The HTTP method is incorrect |  |  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #F55449; margin-bottom: 0.5em; border-radius: 8px;">POST /api/config</div>

*Set a new configuration*

**Request body**: The configuration of netspot (required)

Content (`application/json`): [Config](#config)

**Responses**

|  Code | Description  |
| ------|------------- |
|  **200** | The configuration of netspot is well updated  |
|  **400** | The JSON is invalid (bad format or bad key)  |
|  **405** | The HTTP method is incorrect (it must not occur)  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #29BEB0; margin-bottom: 0.5em; border-radius: 8px;">GET /api/stats/loaded</div>

*Return the statistics currently loaded*

**Responses**

| Code | Description | Content | Return |
|------|-------------|---------|--------|
| **200** | The configuration of netspot | `application/json` | [Loaded](#loaded) |
| **405** | The HTTP method is incorrect |  |  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #29BEB0; margin-bottom: 0.5em; border-radius: 8px;">GET /api/stats/available</div>

*Return the available statistics*

**Responses**

| Code | Description | Content | Return |
|------|-------------|---------|--------|
| **200** | The available statistics | `application/json` | [Available](#available) |
| **405** | The HTTP method is incorrect |  |  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #F55449; margin-bottom: 0.5em; border-radius: 8px;">POST /api/stats/load</div>

*Load new statistics*

**Request body** (required)

Content (`application/json`): [Load](#load)

**Responses**

|  Code | Description  |
| ------|------------- |
|  **200** | The statistics are well loaded  |
|  **400** | The JSON is invalid (bad format or bad key)  |
|  **405** | The HTTP method is incorrect  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #F55449; margin-bottom: 0.5em; border-radius: 8px;">POST /api/stats/unload</div>

*Unload already loaded statistics*

**Request body** (required)

Content (`application/json`): [Unload](#unload)

**Responses**

|  Code | Description  |
| ------|------------- |
|  **200** | The statistics are well unloaded  |
|  **400** | The JSON is invalid (bad format or bad key)  |
|  **405** | The HTTP method is incorrect  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #29BEB0; margin-bottom: 0.5em; border-radius: 8px;">GET /api/stats/values</div>

*Return the current values of the loaded statistics*

**Responses**

| Code | Description | Content | Return |
|------|-------------|---------|--------|
| **200** | The statistics values | `application/json` | [Stat-Values](#stat-values) |
| **405** | The HTTP method is incorrect |  |  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #29BEB0; margin-bottom: 0.5em; border-radius: 8px;">GET /api/stats/status</div>

*Return the DSPOT status of a single statistic*

| Parameters | Type(s) | Details | Required |
|------------|---------|---------|----------|
| `stat` | `string` | Name of the statistic | True |

**Responses**

| Code | Description | Content | Return |
|------|-------------|---------|--------|
| **200** | The DSPOT status of the given statistic | `application/json` | [Status](#status) |
| **405** | The HTTP method is incorrect |  |  |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #29BEB0; margin-bottom: 0.5em; border-radius: 8px;">GET /api/ifaces/available</div>

*Return the available interfaces*

**Responses**

| Code | Description | Content | Return |
|------|-------------|---------|--------|
| **200** | All the interfaces | `application/json` | [Available-Interfaces](#available-interfaces) |

<div style="display: block; width: 100%; color: white; font-size: 1.25em; padding: 0.25em 1em; font-weight: bold; font-family: Liberation Mono; background-color: #F55449; margin-bottom: 0.5em; border-radius: 8px;">POST /api/run</div>

*Perform a running action*

**Request body** (required)

Content (`application/json`): [Run](#run)

**Responses**

|  Code | Description  |
| ------|------------- |
|  **200** | The action related to the command is performed  |
|  **400** | The JSON is invalid (bad format or bad key)  |
|  **405** | The HTTP method is incorrect  |

## Models

### Load
*Statistics to load*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `stats` | `array<string>`, `string` | List of statistics to load |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'stats': ['R_ACK', 'R_SYN', 'PERF'],
}
</code></pre>
        </details>
    </div>
</div>


### Loaded
*Statistics currently loaded*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `loaded` | `array<string>` | List of the statistics currently loaded |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'loaded': ['R_ACK', 'R_ICMP', 'PERF'],
}
</code></pre>
        </details>
    </div>
</div>


### Available
*Available statistics*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `available` | `array<string>` | List of available statistics |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'available': [ 'R_ACK',
                 'R_ICMP',
                 'PERF',
                 'R_DST_SRC',
                 'R_SYN',
                 'AVG_PKT_SIZE'],
}
</code></pre>
        </details>
    </div>
</div>


### Available-Interfaces
*Sniffable interfaces*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `ifaces` | `array<string>` | List of available interfaces |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'ifaces': [ 'eth0',
              'wlp2s0',
              'docker0'],
}
</code></pre>
        </details>
    </div>
</div>


### Unload
*Statistics to unload (possibly all)*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `stats` | `array<string>`, `string` | Statistics to unload |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'stats': ['R_SYN', 'R_ICMP'],
}
</code></pre>
        </details>
    </div>
</div>


### Config
*The configuration*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `device` | `string` | packet source (interface or capture file) |
| `promiscuous` | `boolean` | Promiscuous mode |
| `period` | `string` | Time between two stats computations |
| `output` | `string` | Folder where the data/threshold/anomaly files are stored |
| `file` | `boolean` | Save data/threshold/anomaly to files |
| `influxdb` | `boolean` | Save data/threshold to influxdb |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'device': 'eth0',
  'file': True,
  'influxdb': False,
  'output': '/data',
  'period': '1s',
  'promiscuous': False,
}
</code></pre>
        </details>
    </div>
</div>


### Run
*A running action*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `command` | `string` | Command to send to the server (start, stop, reload, zero) |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'command': 'start',
}
</code></pre>
        </details>
    </div>
</div>


### Stat-Values
*Statistics values*

| Properties | Type(s) | Details |
|------------|---------|---------|
| * | `number` (`double`) | Statistic value |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'R_ACK': 0.012, 'R_SYN': 0.74,
}
</code></pre>
        </details>
    </div>
</div>


### Status
*A DSPOT status*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `drift` | `number` (`double`) | Current average value |
| `n` | `integer` (`int`) | Number of normal observations (not the alarms) |
| `ex_up` | `integer` (`int`) | Current number of up excesses |
| `ex_down` | `integer` (`int`) | Current number of down excesses |
| `Nt_up` | `integer` (`int`) | Total number of up excesses |
| `Nt_down` | `integer` (`int`) | Total number of down excesses |
| `al_up` | `integer` (`int`) | Number of up alarms |
| `al_down` | `integer` (`int`) | Number of down alarms |
| `t_up` | `number` (`double`) | Transitional up threshold |
| `t_down` | `number` (`double`) | Transitional down threshold |
| `z_up` | `number` (`double`) | Up alert threshold |
| `z_down` | `number` (`double`) | Down alert threshold |


<div class="wb-tabs">
	<div class="tabpanels">
		<details id="details-panel1">
			<summary>Example</summary>
			<pre><code class='language-json'>{
  'Nt_down': 0,
  'Nt_up': 120,
  'al_down': 0,
  'al_up': 0,
  'ex_down': 0,
  'ex_up': 120,
  'n': 9950,
  't_up': 2.233,
  'z_up': 4.767,
}
</code></pre>
        </details>
    </div>
</div>


