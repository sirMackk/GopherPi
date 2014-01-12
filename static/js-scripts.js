var deleteItem = function(classSelector, msgItem, url) {

  $(classSelector).on("click", function(e) {
    e.preventDefault();
    e.stopPropagation();
    self = this;
    if (confirm("Are you sure you want to delete this " + msgItem + "?")) {
      media_id = $(this).attr("href");
      $.ajax({
        url: url + media_id,
        type: "DELETE",
        success: function() {
          $(self).parent().parent().remove();
        }
      });
    };
  });
}

