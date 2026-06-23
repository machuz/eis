// CosmicBackdrop (vanilla, for the static Library site) — a port of the Ace
// frontend's WebGL mesh-gradient backdrop. A flowing, painterly aurora rendered
// on a fragment shader: domain-warped fbm churns the field IN PLACE (no rotation
// / no slide), a radial term melts it into the void, a calm centre keeps content
// readable. Per-page accent colours are passed in so each Library section (Ace
// gold, Ideal purple, True teal, Platform copper) gets its own universe.
//
// Usage:  initCosmicBackdrop(canvasEl, ['#d3869b','#b16286','#8a4858','#5b3a6e','#2a3a5e'])
// Colours go warm/bright → deep, used as the gradient ramp. Static if WebGL is
// unavailable or prefers-reduced-motion is set (one frame, no loop).

(function () {
  "use strict";

  var VERT =
    "attribute vec2 a; varying vec2 v_uv;" +
    "void main(){ v_uv = vec2(a.x*0.5+0.5, 0.5-a.y*0.5); gl_Position = vec4(a,0.0,1.0); }";

  var FRAG = [
    "precision highp float; precision highp int;",
    "varying vec2 v_uv;",
    "uniform vec2 u_res; uniform float u_time; uniform vec3 c0,c1,c2,c3,c4;",
    "float hash(vec2 p){ return fract(sin(dot(p,vec2(127.1,311.7)))*43758.5453123); }",
    "float vnoise(vec2 p){ vec2 i=floor(p), f=fract(p);",
    "  float a=hash(i), b=hash(i+vec2(1.,0.)), c=hash(i+vec2(0.,1.)), d=hash(i+vec2(1.,1.));",
    "  vec2 u=f*f*(3.-2.*f); return mix(mix(a,b,u.x), mix(c,d,u.x), u.y); }",
    "float fbm(vec2 p){ float s=0., a=0.55; for(int i=0;i<5;i++){ s+=a*vnoise(p); p=p*2.02+vec2(1.7,9.2); a*=0.5; } return s; }",
    "vec3 ramp(float x){ x=clamp(x,0.,1.); vec3 col=c0;",
    "  col=mix(col,c1,smoothstep(0.0,0.28,x)); col=mix(col,c2,smoothstep(0.22,0.5,x));",
    "  col=mix(col,c3,smoothstep(0.46,0.74,x)); col=mix(col,c4,smoothstep(0.7,1.0,x)); return col; }",
    "void main(){ vec2 uv=v_uv; vec2 asp=vec2(u_res.x/u_res.y,1.0);",
    "  vec2 p=(uv-0.5)*asp*2.2; float t=u_time*0.05;",
    "  vec2 q=vec2(fbm(p+t), fbm(p+vec2(5.2,1.3)-t));",
    "  vec2 r=vec2(fbm(p+1.8*q+vec2(1.7,9.2)+0.7*t), fbm(p+1.8*q+vec2(8.3,2.8)-0.6*t));",
    "  float f=fbm(p+2.2*r);",
    "  float hpos = uv.x*0.85 + (f-0.5)*0.7 + (r.x-0.5)*0.35;",
    "  vec3 col=ramp(hpos); col *= 0.7 + 0.6*f;",
    "  vec2 cc=(uv-vec2(0.5,0.44))*asp; float d=length(cc);",
    "  float melt=1.0 - smoothstep(0.34,0.82,d);",
    "  float calm=smoothstep(0.0,0.36,d); col*=mix(0.55,1.0,calm);",
    "  col*=melt; float st=hash(floor(uv*asp*420.0)); col+=step(0.9975,st)*0.4*melt;",
    "  gl_FragColor=vec4(col,1.0); }",
  ].join("\n");

  function hexToRGB(h) {
    h = h.replace("#", "");
    if (h.length === 3) h = h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
    return [parseInt(h.slice(0, 2), 16) / 255, parseInt(h.slice(2, 4), 16) / 255, parseInt(h.slice(4, 6), 16) / 255];
  }

  function compile(gl, type, src) {
    var s = gl.createShader(type);
    gl.shaderSource(s, src);
    gl.compileShader(s);
    if (!gl.getShaderParameter(s, gl.COMPILE_STATUS)) console.error("cosmic shader:", gl.getShaderInfoLog(s));
    return s;
  }

  window.initCosmicBackdrop = function (canvas, colors) {
    if (!canvas) return;
    var gl = canvas.getContext("webgl", { antialias: false, alpha: false });
    if (!gl) {
      canvas.style.background =
        "radial-gradient(60% 60% at 40% 38%, " + colors[1] + "44, transparent 70%), #0a0e17";
      return;
    }
    var prog = gl.createProgram();
    gl.attachShader(prog, compile(gl, gl.VERTEX_SHADER, VERT));
    gl.attachShader(prog, compile(gl, gl.FRAGMENT_SHADER, FRAG));
    gl.linkProgram(prog);
    if (!gl.getProgramParameter(prog, gl.LINK_STATUS)) {
      console.error("cosmic link:", gl.getProgramInfoLog(prog));
      return;
    }
    gl.useProgram(prog);

    var buf = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buf);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 3, -1, -1, 3]), gl.STATIC_DRAW);
    var loc = gl.getAttribLocation(prog, "a");
    gl.enableVertexAttribArray(loc);
    gl.vertexAttribPointer(loc, 2, gl.FLOAT, false, 0, 0);

    var uRes = gl.getUniformLocation(prog, "u_res");
    var uTime = gl.getUniformLocation(prog, "u_time");
    for (var i = 0; i < 5; i++) gl.uniform3fv(gl.getUniformLocation(prog, "c" + i), hexToRGB(colors[i]));
    gl.clearColor(0, 0, 0, 1);

    var reduce = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;

    function resize() {
      var dpr = Math.min(window.devicePixelRatio || 1, 1.5);
      var w = Math.max(1, Math.round(canvas.clientWidth * dpr));
      var h = Math.max(1, Math.round(canvas.clientHeight * dpr));
      if (canvas.width !== w || canvas.height !== h) { canvas.width = w; canvas.height = h; gl.viewport(0, 0, w, h); }
      gl.uniform2f(uRes, canvas.width, canvas.height);
    }
    var raf = 0, start = performance.now();
    function draw(now) {
      resize();
      gl.uniform1f(uTime, reduce ? 0 : (now - start) / 1000);
      gl.clear(gl.COLOR_BUFFER_BIT);
      gl.drawArrays(gl.TRIANGLES, 0, 3);
      raf = reduce || document.visibilityState !== "visible" ? 0 : requestAnimationFrame(draw);
    }
    draw(start);
    document.addEventListener("visibilitychange", function () {
      if (!reduce && document.visibilityState === "visible" && !raf) raf = requestAnimationFrame(draw);
    });
    window.addEventListener("resize", function () { if (!raf) draw(performance.now()); });
  };
})();
