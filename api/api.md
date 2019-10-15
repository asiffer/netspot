### Models

#### Load
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


#### Loaded
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


#### Available
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


#### Unload
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


#### Config
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


#### Run
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


#### Status
*A DSPOT status*

| Properties | Type(s) | Details |
|------------|---------|---------|
| `drift` | `number` (`float64`) | Current average value |
| `n` | `integer` (`int`) | Number of normal observations (not the alarms) |
| `ex_up` | `integer` (`int`) | Current number of up excesses |
| `ex_down` | `integer` (`int`) | Current number of down excesses |
| `Nt_up` | `integer` (`int`) | Total number of up excesses |
| `Nt_down` | `integer` (`int`) | Total number of down excesses |
| `al_up` | `integer` (`int`) | Number of up alarms |
| `al_down` | `integer` (`int`) | Number of down alarms |
| `t_up` | `number` (`float64`) | Transitional up threshold |
| `t_down` | `number` (`float64`) | Transitional down threshold |
| `z_up` | `number` (`float64`) | Up alert threshold |
| `z_down` | `number` (`float64`) | Down alert threshold |


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


