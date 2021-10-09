var app = new Vue({
  el: "#app",
  data: {
   x: 42,
   y: 55,
   scale: 25,
   seed: 0, 
  },
  methods: {
    randomseed:  function() {
      this.seed = Math.round(Math.random() * Math.pow(2,63))
    },
  },
  computed: {
    svgurl: function() {
      return "/api/maze/"+
        this.x+"x"+this.y+"/" + this.seed + "?s="+this.scale
    },
  },
  mounted: function() {
    this.randomseed()
  },
})
