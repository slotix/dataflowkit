DownCount
=========

jQuery countdown plugin that accounts for timezone and daylight.

#Usage

```JS
$('.countdown').downCount({
    date: '08/27/2013 12:00:00',
    offset: -5
}, function () {
    alert('WOOT WOOT, done!');
});
```

#Options
Option | Description
---|---
date | Target date, ex `08/27/2013 12:00:00`
offset | UTC Standard Timezone offset

You can also append a callback function which is called when countdown finishes.