$(function () {

  /* ==================================================================
   Default Demo  
   ================================================================== */

  $('#summernote-basic-demo').summernote ({ 
    height: 150 
  })



  /* ==================================================================
   Arimode Demo 
   ================================================================== */

  $('#summernote-airmode-demo').summernote({
    airMode: true
  })



  /* ==================================================================
   Hint Demo 
   ================================================================== */

  $("#summernote-hint-demo").summernote({
    height: 150,
    toolbar: false,
    placeholder: 'type with apple, orange, watermelon and lemon',
    hint: {
      words: ['apple', 'orange', 'watermelon', 'lemon'],
      match: /\b(\w{1,})$/,
      search: function (keyword, callback) {
        callback($.grep(this.words, function (item) {
          return item.indexOf(keyword) === 0
        }))
      }
    }
  })

  $.ajax({
    url: 'https://api.github.com/emojis',
    async: false 
  }).then(function(data) {
    window.emojis = Object.keys(data)
    window.emojiUrls = data 
  })



  /* ==================================================================
   Emoji Demo 
   ================================================================== */

  $("#summernote-emoji-demo").summernote({
    height: 150,
    toolbar: false,
    placeholder: 'type starting with : and any alphabet',
    hint: {
      match: /:([\-+\w]+)$/,
      search: function (keyword, callback) {
        callback($.grep(emojis, function (item) {
          return item.indexOf(keyword)  === 0
        }))
      },
      template: function (item) {
        var content = emojiUrls[item]
        return '<img src="' + content + '" width="20" /> :' + item + ':'
      },
      content: function (item) {
        var url = emojiUrls[item]
        if (url) {
          return $('<img />').attr('src', url).css('width', 20)[0]
        }
        return ''
      }
    }
  })



  /* ==================================================================
   @ Mentions Demo 
   ================================================================== */

  $("#summernote-mention-demo").summernote({
    height: 150,
    toolbar: false,
    placeholder: 'type starting with @',
    hint: {
      mentions: ['jayden', 'sam', 'alvin', 'david'],
      match: /\B@(\w*)$/,
      search: function (keyword, callback) {
        callback($.grep(this.mentions, function (item) {
          return item.indexOf(keyword) == 0
        }))
      },
      content: function (item) {
        return '@' + item
      }    
    }
  })

})