# Very simple jQuery color picker

Yet another jQuery color picker. This plugin is unobtrusive and integrates well with Twitter Bootstrap (it works just fine without).
The source code only requires jQuery and is about 200 lines of JavaScript and 100 lines of CSS.

See the [online demo](http://plnkr.co/edit/VVclW0?p=preview).

![simplecolorpicker-inline.png](http://img11.hostingpics.net/pics/75504320131025121603ColorpickerforjQuery.png)

![simplecolorpicker-picker.png](http://img11.hostingpics.net/pics/71709820131025122115ColorpickerforjQuery.png)

## How to use

Create a HTML select:

```HTML
<select name="colorpicker">
  <option value="#7bd148">Green</option>
  <option value="#5484ed">Bold blue</option>
  <option value="#a4bdfc">Blue</option>
  <option value="#46d6db">Turquoise</option>
  <option value="#7ae7bf">Light green</option>
  <option value="#51b749">Bold green</option>
  <option value="#fbd75b">Yellow</option>
  <option value="#ffb878">Orange</option>
  <option value="#ff887c">Red</option>
  <option value="#dc2127">Bold red</option>
  <option value="#dbadff">Purple</option>
  <option value="#e1e1e1">Gray</option>
</select>
```

add the plugin files:

```HTML
<script src="//ajax.googleapis.com/ajax/libs/jquery/1.10.2/jquery.js"></script>
<script src="jquery.simplecolorpicker.js"></script>

<link rel="stylesheet" href="jquery.simplecolorpicker.css">
```

then call the plugin:

```JavaScript
$('select[name="colorpicker"]').simplecolorpicker();
$('select[name="colorpicker"]').simplecolorpicker('selectColor', '#7bd148');
$('select[name="colorpicker"]').simplecolorpicker('destroy');
```

and pass some options if needed:

```JavaScript
$('select[name="colorpicker"]').simplecolorpicker({
  picker: true
}).on('change', function() {
  $(document.body).css('background-color', $('select[name="colorpicker"]').val());
});
```

### Options

- theme: font to use for the ok/check mark (default: `''`), available themes: [`regularfont`](https://github.com/tkrotoff/jquery-simplecolorpicker/blob/master/jquery.simplecolorpicker-regularfont.css), [`fontawesome`](https://github.com/tkrotoff/jquery-simplecolorpicker/blob/master/jquery.simplecolorpicker-fontawesome.css), [`glyphicons`](https://github.com/tkrotoff/jquery-simplecolorpicker/blob/master/jquery.simplecolorpicker-glyphicons.css)
- picker: show the colors inside a picker instead of inline (default: `false`)
- pickerDelay: show and hide animation delay in milliseconds (default: `0`)

## Browser support

Simplecolorpicker supports all modern browsers starting with Internet Explorer 8 included.
It is recommended to not use any font theme with IE8.

## HTML5 new color input

HTML5 provides a new input to select colors. Its implementation inside modern browsers is still lacking.
The new color input does not provide any option to customize the color list and
most of the time the widget is less user-friendly than the one provided by simplecolorpicker.

See http://slides.html5rocks.com/#new-form-types

See http://dev.w3.org/html5/markup/input.color.html#input.color

## Bower

```
bower install jquery-simplecolorpicker
```

## AngularJS directive

See [simplecolorpicker directive](http://plnkr.co/edit/rKM3QWXDC3vGVPe3QFWV?p=preview).
If you find a solution for the `setTimeout()` hack, please tell me.

Here [another directive](http://plnkr.co/edit/zlP0RSH3m0ghsefHeaLI?p=preview) written by [KGZM](https://github.com/KGZM) that re-implements simplecolorpicker.
For the explanations, read this [Google Groups discussion](https://groups.google.com/d/topic/angular/nBZsvLOZxvI/discussion).

## Ruby on Rails

A gem is available at https://github.com/tkrotoff/jquery-simplecolorpicker-rails

## Copyright and license

Licensed under the MIT license.
Copyright (C) 2012-2013 Tanguy Krotoff
