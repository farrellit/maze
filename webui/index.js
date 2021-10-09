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
      this.seed = Math.round(Math.random() * Math.pow(2,64))
    },
  },
  computed: {
    svgurl: function() {
      return "/api/maze?x="+
        this.x+"&y="+this.y+"&s="+this.scale+"&seed="+this.seed+"&"
    },
  },
  mounted: function() {
    this.randomseed()
  },
})
