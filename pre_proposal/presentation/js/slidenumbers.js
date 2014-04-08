var slides = $(".step").toArray();
var totalSlides = slides.length;

for(var i=0; i<totalSlides; i++){
  $(slides[i]).prepend(
  '<div class="counter"> ' + (i+1) +"/" + totalSlides + "</div>"
  );
}
