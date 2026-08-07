package main

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const demoUIHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="description" content="Run DeltaSignal's governed HUT research-memory proof after exploring the flagship Apple evidence workflow.">
  <title>DeltaSignal Gemini AI Agent Demo</title>
  <style>
    :root{color-scheme:dark;--bg:#05070b;--panel:#101722;--line:rgba(255,255,255,.14);--text:#f7fbff;--muted:#b9c7d8;--green:#5df58b;--blue:#72d9ff;--amber:#f8d998;--orange:#ff7a3d;--gold:#d6ad5c;--violet:#caa8ff;--pink:#ff8ad8;--mono:"SF Mono","JetBrains Mono","Cascadia Code",ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}
    *{box-sizing:border-box}
    html{width:100%;height:100%;overflow:hidden;background:#05070b}
    body{margin:0;width:100%;height:100%;min-height:100vh;overflow:hidden;background:linear-gradient(135deg,#04070c 0%,#08111a 47%,#07120d 100%);color:var(--text);font:15px/1.4 Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;-webkit-font-smoothing:antialiased;text-rendering:geometricPrecision}
    demo-app{display:block;width:100vw;height:100vh;height:100dvh;min-height:100vh;min-height:100dvh;overflow:hidden;background:radial-gradient(circle at 76% 12%,rgba(93,245,139,.12),transparent 28%),linear-gradient(135deg,#04070c 0%,#08111a 48%,#06110d 100%)}
    body::before{content:"";position:fixed;inset:0;pointer-events:none;background-image:linear-gradient(rgba(255,255,255,.024) 1px,transparent 1px),linear-gradient(90deg,rgba(255,255,255,.024) 1px,transparent 1px);background-size:44px 44px;mask-image:linear-gradient(to bottom,rgba(0,0,0,.75),transparent 74%)}
    main{height:100dvh;max-height:100dvh;max-width:1320px;margin:0 auto;padding:14px;position:relative;display:grid;grid-template-rows:auto minmax(0,1fr) auto;gap:12px;overflow:hidden}
    header,.panel,.miniCard{border:1px solid var(--line);background:linear-gradient(180deg,rgba(18,28,43,.98),rgba(7,12,20,.98));border-radius:16px;box-shadow:0 18px 54px rgba(0,0,0,.34),inset 0 1px 0 rgba(255,255,255,.075)}
    header{display:grid;grid-template-columns:minmax(0,1fr) auto;gap:14px;align-items:center;padding:13px 15px;border-color:rgba(214,173,92,.42);overflow:hidden;position:relative}
    header::after{content:"";position:absolute;right:-90px;top:-150px;width:320px;height:320px;border-radius:50%;background:radial-gradient(circle,rgba(214,173,92,.24),rgba(93,245,139,.12),transparent 64%);filter:blur(2px)}
    .brandRow{display:flex;align-items:center;gap:13px;position:relative;z-index:1;min-width:0}
    .brandLink{display:flex;align-items:center;gap:13px;min-width:0;color:inherit;text-decoration:none}
    .brandLink:hover .logo{box-shadow:0 0 0 1px rgba(255,255,255,.22),0 0 34px rgba(214,173,92,.52)}
    .logo{width:50px;height:50px;border-radius:14px;object-fit:cover;box-shadow:0 0 0 1px rgba(255,255,255,.18),0 0 30px rgba(214,173,92,.34);flex:0 0 auto}
    h1{font-size:clamp(30px,3.2vw,45px);line-height:.96;margin:3px 0 5px;max-width:900px;letter-spacing:0}
    .proofTitle{opacity:0;transform:translateY(16px) scale(.985);filter:blur(7px);animation:proofTitleIn 1.05s cubic-bezier(.18,.72,.16,1) .18s forwards;text-shadow:0 0 26px rgba(255,255,255,.12)}
    .stageReveal{opacity:0;transform:translateY(14px);filter:blur(5px);animation:stageIn .82s cubic-bezier(.2,.72,.18,1) forwards;animation-delay:var(--delay,.6s)}
    h2{margin:0;color:var(--green);font-size:19px;line-height:1.08}
    p{color:var(--muted);margin:0}
    .eyebrow{color:var(--gold);text-transform:uppercase;letter-spacing:.12em;font-weight:950;font-size:12px}
    .entryName{display:inline-flex;margin-top:4px;color:#f7fbff;font-weight:950;font-size:15px;letter-spacing:.02em;text-transform:uppercase}
    .headerControls{display:grid;gap:8px;justify-items:end;position:relative;z-index:1;min-width:540px}
    .status{display:inline-flex;align-items:center;gap:8px;border:1px solid rgba(214,173,92,.38);border-radius:999px;padding:8px 12px;color:var(--gold);font-weight:950;background:rgba(214,173,92,.08)}
    .headerInputGrid{width:540px;display:grid;grid-template-columns:1.05fr 1.35fr .95fr;gap:8px}
    .headerInputGrid label{font-size:10px;margin-bottom:4px;text-align:left}
    .headerInputGrid input{height:33px;padding:7px 10px}
    .dot{width:9px;height:9px;border-radius:999px;background:var(--green);box-shadow:0 0 18px rgba(93,245,139,.7)}
    .appGrid{display:grid;grid-template-columns:38.2fr 61.8fr;gap:12px;min-height:0;height:100%;overflow:hidden}
    .panel{padding:12px;min-height:0;overflow:hidden}
    .controlPanel{display:grid;grid-template-rows:auto auto auto auto minmax(0,1fr);gap:10px;border-color:rgba(214,173,92,.30);background:linear-gradient(180deg,rgba(15,24,35,.98),rgba(7,12,17,.99));box-shadow:0 18px 54px rgba(0,0,0,.34),0 0 0 1px rgba(214,173,92,.045),inset 0 1px 0 rgba(255,255,255,.075)}
    .responsePanel{display:grid;grid-template-rows:auto minmax(0,1fr);gap:12px;border-color:rgba(114,217,255,.62);background:radial-gradient(circle at 100% 0%,rgba(114,217,255,.18),transparent 34%),linear-gradient(180deg,rgba(6,31,57,.99),rgba(1,12,27,.995));box-shadow:0 24px 70px rgba(0,0,0,.48),0 0 0 2px rgba(114,217,255,.16),0 0 54px rgba(114,217,255,.16),inset 0 1px 0 rgba(255,255,255,.11)}
    .responsePanel h2{color:#8ee5ff;text-shadow:0 0 18px rgba(114,217,255,.34)}
    .responsePanel .pill{border-color:rgba(114,217,255,.42);color:#aeeeff;background:rgba(114,217,255,.11)}
    .panelHeader{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-bottom:6px}
    .panelHeader p{font-size:13px}
    .miniCard{padding:10px;border-radius:13px;background:linear-gradient(180deg,rgba(255,255,255,.065),rgba(255,255,255,.026))}
    .miniCard b{display:block;color:var(--text)}
    .miniCard span{display:block;color:var(--muted);font-size:12px;margin-top:4px}
    .runGrid{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:7px}
    .inputPair{display:grid;grid-template-columns:1fr 1fr;gap:8px}
    label{display:block;color:var(--amber);font-weight:900;font-size:11px;text-transform:uppercase;letter-spacing:.09em;margin:0 0 5px}
    input,textarea{width:100%;border:1px solid rgba(114,217,255,.30);background:#07101d;color:var(--text);border-radius:10px;padding:8px 10px;font:13px var(--mono)}
    textarea{min-height:86px;resize:vertical}
    button{position:relative;overflow:hidden;border:1px solid rgba(93,245,139,.36);background:linear-gradient(135deg,rgba(93,245,139,.25),rgba(114,217,255,.14));color:var(--text);border-radius:10px;padding:9px 8px;font-weight:950;font-size:12px;cursor:pointer;box-shadow:0 10px 28px rgba(0,0,0,.22);transition:transform .14s ease,border-color .14s ease,box-shadow .14s ease}
    button .btnIcon{display:inline-flex;align-items:center;justify-content:center;margin-right:5px;color:#ffe2a3;filter:drop-shadow(0 0 8px rgba(214,173,92,.34));vertical-align:-.18em}
    button .btnIcon svg{width:15px;height:15px;stroke:currentColor;stroke-width:2.4;fill:none;stroke-linecap:round;stroke-linejoin:round}
    button:hover{border-color:var(--green);transform:translateY(-1px);box-shadow:0 16px 42px rgba(93,245,139,.13)}
    button.secondary{border-color:rgba(114,217,255,.3);background:rgba(114,217,255,.08)}
    button.running{border-color:var(--green);background:linear-gradient(135deg,rgba(93,245,139,.36),rgba(114,217,255,.22));box-shadow:0 0 0 2px rgba(93,245,139,.14),0 0 34px rgba(93,245,139,.24);animation:buttonPulse 1s ease-in-out infinite}
    button.running::after{content:"";position:absolute;inset:-45%;background:linear-gradient(90deg,transparent,rgba(255,255,255,.26),transparent);transform:translateX(-70%);animation:buttonSweep 1.15s linear infinite}
    pre{white-space:pre-wrap;word-break:break-word;margin:0;background:linear-gradient(180deg,#041424,#02070e);border:1px solid rgba(114,217,255,.42);border-radius:17px;padding:18px;height:100%;overflow:auto;color:#e9fbff;font:15px/1.62 var(--mono);letter-spacing:0;-webkit-font-smoothing:antialiased;text-rendering:geometricPrecision;box-shadow:inset 0 1px 0 rgba(255,255,255,.08),0 0 44px rgba(114,217,255,.075),0 0 0 1px rgba(0,0,0,.24)}
    pre::selection{background:rgba(114,217,255,.28)}
    .resp-title{color:var(--green);font-weight:950}
    .resp-status{color:var(--blue);font-weight:950}
    .resp-error{color:#ff6b75;font-weight:950}
    .json-key{color:#7ddcff;font-weight:850}
    .json-string{color:#f7d48a}
    .json-number{color:#caa8ff}
    .json-bool{color:#5df58b;font-weight:850}
    .json-null{color:#ff94a8;font-weight:850}
    .json-punc{color:#7d8ca7}
    .json-brace{color:#d8f3ff;font-weight:900}
    .json-colon{color:#d6ad5c}
    .traceBox{display:block;margin:0 0 14px;padding:12px 14px;border:1px solid rgba(93,245,139,.26);border-radius:14px;background:linear-gradient(180deg,rgba(93,245,139,.09),rgba(114,217,255,.05));box-shadow:0 0 30px rgba(93,245,139,.07)}
    .traceTitle{display:flex;align-items:center;gap:8px;color:var(--green);font-weight:950;margin-bottom:8px}
    .traceDot{width:9px;height:9px;border-radius:999px;background:var(--green);box-shadow:0 0 18px rgba(93,245,139,.82);animation:tracePulse 1s ease-in-out infinite}
    .traceGrid{display:grid;grid-template-columns:128px minmax(0,1fr);gap:5px 12px;color:#dff6ff}
    .traceGrid b{color:#8fa0ba;font-weight:850}
    .traceGrid span{color:#f7fbff}
    .traceGrid .ok{color:var(--green);font-weight:950}
    .traceGrid .route{color:var(--blue);font-weight:850}
    .streamCursor{display:inline-block;width:8px;height:1.05em;margin-left:3px;vertical-align:-.15em;background:var(--green);box-shadow:0 0 12px rgba(93,245,139,.85);animation:cursorBlink .78s steps(2,end) infinite}
    .responseTools{display:flex;align-items:center;gap:7px;flex-wrap:wrap}
    .responseTools button{padding:7px 10px;font-size:12px;border-color:rgba(114,217,255,.3);background:rgba(114,217,255,.075)}
    .responseTools button.active{border-color:var(--green);box-shadow:0 0 26px rgba(93,245,139,.22);animation:buttonPulse 1s ease-in-out infinite}
    .responseTools button.renderButton{border-color:rgba(214,173,92,.5);background:linear-gradient(135deg,rgba(214,173,92,.22),rgba(114,217,255,.08));color:#ffe2a3}
    .richRender{display:grid;gap:14px;white-space:normal;font-family:Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:#f7fbff}
    .richHero{display:grid;gap:10px;border:1px solid rgba(214,173,92,.48);border-radius:18px;padding:16px;background:linear-gradient(135deg,rgba(214,173,92,.17),rgba(93,245,139,.055),rgba(114,217,255,.06));box-shadow:0 20px 54px rgba(0,0,0,.28),inset 0 1px 0 rgba(255,255,255,.08)}
    .richKicker{color:#d3a63a;text-transform:uppercase;letter-spacing:.11em;font-weight:950;font-size:12px}
    .richTitle{font-size:30px;line-height:1.02;font-weight:1000;color:#fff;margin:0;text-shadow:0 1px 18px rgba(0,0,0,.52)}
    .richSummary{color:#e8f2ff;font-size:17px;line-height:1.34;font-weight:760}
    .richGrid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:10px}
    .richCard{border:1px solid rgba(114,217,255,.24);border-radius:15px;padding:13px;background:linear-gradient(180deg,rgba(114,217,255,.09),rgba(255,255,255,.028));box-shadow:inset 0 1px 0 rgba(255,255,255,.06)}
    .richCard.gold{border-color:rgba(214,173,92,.42);background:linear-gradient(180deg,rgba(214,173,92,.12),rgba(255,255,255,.026))}
    .richCard.green{border-color:rgba(93,245,139,.36);background:linear-gradient(180deg,rgba(93,245,139,.10),rgba(255,255,255,.026))}
    .richCard b{display:block;color:#90e9ff;text-transform:uppercase;letter-spacing:.08em;font-size:11px;margin-bottom:7px}
    .richCard span,.richCard li{color:#f4fbff;font-size:14px;line-height:1.38;font-weight:720}
    .richCard code{font-family:var(--mono);color:#f7d48a;font-weight:850;word-break:break-word}
    .richList{display:grid;gap:7px;margin:0;padding:0;list-style:none}
    .richSection{display:grid;gap:10px}
    .richSection h3{margin:0;color:#5df58b;font-size:20px;line-height:1.12}
    .richSection p{color:#d8e5f5;font-size:15px;line-height:1.42;font-weight:680}
    .richTable{display:grid;gap:6px}
    .richRow{display:grid;grid-template-columns:150px minmax(0,1fr);gap:10px;border:1px solid rgba(255,255,255,.09);border-radius:11px;padding:8px 10px;background:rgba(0,0,0,.18)}
    .richRow b{color:#8fa0ba;font-size:12px;text-transform:uppercase;letter-spacing:.08em}
    .richRow span{color:#f7fbff;font-size:14px;font-weight:730;word-break:break-word}
    .richPre{white-space:pre-wrap;word-break:break-word;border:1px solid rgba(114,217,255,.25);border-radius:14px;padding:13px;background:#03111f;color:#dff6ff;font:13px/1.45 var(--mono);max-height:360px;overflow:auto}
    .articleStart,.overviewStart{display:grid;grid-template-columns:minmax(0,1fr);gap:14px;color:var(--text);white-space:normal}
    .objectiveBox{display:block;border:1px solid rgba(214,173,92,.38);border-radius:14px;padding:13px 15px;background:linear-gradient(135deg,rgba(214,173,92,.14),rgba(114,217,255,.06));box-shadow:0 0 34px rgba(214,173,92,.08)}
    .objectiveBox b{display:block;color:#ffe2a3;text-transform:uppercase;letter-spacing:.09em;font-size:13px;margin-bottom:6px}
    .objectiveBox span{display:block;color:#f7fbff;font:720 16px/1.45 Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif}
    .overviewGrid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:10px}
    .overviewGrid .miniCard{border-color:rgba(114,217,255,.24);background:linear-gradient(180deg,rgba(114,217,255,.085),rgba(93,245,139,.035))}
    .overviewGrid b{color:#dff6ff}
    .overviewGrid span{font-size:13px;line-height:1.36}
    .captionBar{display:block;position:relative;z-index:30;border:2px solid rgba(248,217,152,.66);border-radius:16px;padding:13px 18px;background:linear-gradient(135deg,#07090e,#10141c 48%,#22190b);box-shadow:0 -20px 64px rgba(0,0,0,.58),0 0 0 1px rgba(0,0,0,.75),0 0 38px rgba(214,173,92,.16),inset 0 1px 0 rgba(255,255,255,.12);min-height:88px;overflow:hidden}
    .captionText{display:block;color:#d3a63a;font:950 25px/1.08 "Arial Narrow","Roboto Condensed","Aptos Narrow","Inter Tight",Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;letter-spacing:.01em;max-height:3.45em;overflow:hidden;text-shadow:0 2px 0 rgba(0,0,0,.78),0 0 18px rgba(211,166,58,.22)}
    .captionCursor{display:inline-block;width:10px;height:1em;margin-left:5px;vertical-align:-.12em;background:#d3a63a;box-shadow:0 0 18px rgba(211,166,58,.92);animation:cursorBlink .78s steps(2,end) infinite}
    .watermark{position:fixed;right:22px;bottom:18px;z-index:20;color:rgba(247,251,255,.44);font:850 10px/1.2 Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;letter-spacing:.09em;text-transform:uppercase;text-shadow:0 1px 12px rgba(0,0,0,.72);pointer-events:none}
    .articleHero{display:grid;grid-template-columns:1fr;gap:10px}
    .articleHero img{width:100%;max-height:420px;object-fit:cover;object-position:top left;border-radius:14px;border:1px solid rgba(214,173,92,.35);box-shadow:0 20px 60px rgba(0,0,0,.42)}
    .substackURL{display:block;border:1px solid rgba(214,173,92,.35);border-radius:12px;padding:10px 12px;background:rgba(214,173,92,.075);color:#ffe2a3;font:900 14px/1.25 var(--mono);letter-spacing:0;white-space:normal;word-break:break-all}
    .articleMeta{display:grid;grid-template-columns:1.1fr .9fr;gap:10px}
    .articleMeta .miniCard{border-color:rgba(93,245,139,.24);background:linear-gradient(180deg,rgba(93,245,139,.08),rgba(114,217,255,.035))}
    .articleMeta code{color:var(--amber);font-family:var(--mono);font-weight:900}
    .articleLink{display:inline-flex;align-items:center;justify-content:center;width:max-content;border:1px solid rgba(214,173,92,.42);border-radius:999px;padding:8px 12px;color:var(--gold);text-decoration:none;font-weight:950;background:rgba(214,173,92,.08)}
    .flow{display:grid;grid-template-columns:repeat(5,1fr);gap:6px;position:relative;padding:8px;border:1px solid rgba(114,217,255,.16);border-radius:14px;background:rgba(0,0,0,.2);overflow:hidden}
    .flow::before{content:"";position:absolute;left:28px;right:28px;top:50%;height:2px;background:linear-gradient(90deg,rgba(93,245,139,.15),rgba(114,217,255,.8),rgba(93,245,139,.15));box-shadow:0 0 22px rgba(114,217,255,.35)}
    .flow::after{content:"";position:absolute;left:24px;top:calc(50% - 7px);width:14px;height:14px;border-radius:50%;background:var(--green);box-shadow:0 0 20px rgba(93,245,139,.82);animation:packet 5.8s linear infinite}
    .flow div{position:relative;z-index:1;border:1px solid rgba(255,255,255,.12);border-radius:11px;padding:6px;text-align:center;color:var(--muted);font-size:10px;background:rgba(7,13,22,.86)}
    .flow b{display:block;color:var(--text);font-size:12px}
    .metric strong{display:block;color:var(--green);font-size:24px;line-height:1}
    .metric span{display:block;color:var(--muted);font-size:10px;font-weight:800;margin-top:3px}
    .inputSpacer{height:6px}
    .inspectCard{border-color:rgba(214,173,92,.25);background:rgba(214,173,92,.06)}
    .actionCard{align-self:stretch;display:grid;grid-template-rows:auto minmax(0,1fr);gap:7px;border-color:rgba(214,173,92,.36);background:linear-gradient(180deg,rgba(214,173,92,.12),rgba(93,245,139,.045));box-shadow:inset 0 1px 0 rgba(255,255,255,.08),0 0 34px rgba(214,173,92,.08)}
    .actionCard b{display:flex;align-items:center;gap:8px;color:#ffe2a3;text-transform:uppercase;letter-spacing:.08em;font-size:14px}
    .actionCard b::before{content:"";width:8px;height:8px;border-radius:999px;background:var(--gold);box-shadow:0 0 16px rgba(214,173,92,.7)}
    .miniCard .actionText{display:block;color:#f4fbff;font-size:20px;line-height:1.24;min-height:114px;font-weight:850;text-shadow:0 1px 14px rgba(0,0,0,.44)}
    .actionCursor{display:inline-block;width:7px;height:1em;margin-left:3px;vertical-align:-.15em;background:var(--gold);box-shadow:0 0 12px rgba(214,173,92,.8);animation:cursorBlink .78s steps(2,end) infinite}
    .pill{display:inline-flex;align-items:center;gap:6px;border:1px solid rgba(93,245,139,.24);border-radius:999px;padding:6px 8px;color:var(--green);font-weight:950;font-size:11px;background:rgba(93,245,139,.07)}
    .topline{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:7px}
    .topline .miniCard{min-height:60px}
    .mono{font-family:var(--mono)}
    @keyframes packet{0%{transform:translateX(0);opacity:.2}8%{opacity:1}92%{opacity:1}100%{transform:translateX(calc(100vw - 120px));opacity:.2}}
    @keyframes tracePulse{0%,100%{transform:scale(.82);opacity:.55}50%{transform:scale(1.25);opacity:1}}
    @keyframes cursorBlink{0%,45%{opacity:1}46%,100%{opacity:.12}}
    @keyframes buttonPulse{0%,100%{transform:translateY(0) scale(1)}50%{transform:translateY(-1px) scale(1.025)}}
    @keyframes buttonSweep{0%{transform:translateX(-70%) rotate(10deg)}100%{transform:translateX(70%) rotate(10deg)}}
    @keyframes proofTitleIn{0%{opacity:0;transform:translateY(18px) scale(.975);filter:blur(9px)}62%{opacity:1;filter:blur(0)}100%{opacity:1;transform:translateY(0) scale(1);filter:blur(0)}}
    @keyframes stageIn{0%{opacity:0;transform:translateY(16px);filter:blur(6px)}74%{opacity:1;filter:blur(0)}100%{opacity:1;transform:translateY(0);filter:blur(0)}}
    @media(max-height:760px){
      header p{display:none}
      .logo{width:40px;height:40px}
      h1{font-size:clamp(23px,3vw,36px)}
      .eyebrow{font-size:10px}
      .entryName{font-size:12px}
      .headerControls{min-width:470px}
      .headerInputGrid{width:470px}
      .captionBar{min-height:78px;padding:11px 14px}
      .captionText{font-size:22px;line-height:1.08}
      .flow{display:none}
      .inspectCard{display:none}
      .actionCard{display:grid}
      .miniCard .actionText{font-size:17px;line-height:1.2;min-height:82px}
      .panelHeader p{display:none}
      .topline .miniCard{min-height:46px}
      .metric strong{font-size:20px}
      .inputPair{grid-template-columns:1fr 1fr}
      pre{font-size:14px;line-height:1.54}
    }
    @media(max-width:900px){
      html,body{width:100%;max-width:100vw;height:auto;min-height:100%;overflow-x:hidden;overflow-y:auto}
      demo-app{display:block;width:100%;max-width:100vw;height:auto;min-height:100dvh;overflow:visible}
      main{width:100%;max-width:100vw;height:auto;max-height:none;min-height:100dvh;padding:10px;overflow:visible}
      .appGrid,header{grid-template-columns:minmax(0,1fr)}
      .appGrid,header,.panel,.brandRow,.brandLink,.headerControls,.headerInputGrid,.panelHeader,.topline,.flow,.runGrid,.responseTools,pre{min-width:0;max-width:100%}
      header{overflow:hidden}
      header::after{display:none}
      .brandRow,.brandLink{width:100%;min-width:0;align-items:flex-start}
      .brandLink>div{min-width:0}
      h1{font-size:clamp(28px,9vw,40px);line-height:1.02;overflow-wrap:anywhere}
      header p{display:block;font-size:13px}
      .headerControls{width:100%;min-width:0;justify-items:stretch}
      .headerInputGrid{width:100%;grid-template-columns:minmax(0,1fr)}
      .panel{width:100%;overflow:visible}
      .controlPanel,.responsePanel{grid-template-rows:auto}
      .runGrid,.flow,.metrics,.topline{grid-template-columns:repeat(2,minmax(0,1fr))}
      .brandRow{align-items:flex-start}
      .logo{width:48px;height:48px}
      .captionBar{min-height:0}
      .captionText{font-size:19px;max-height:none}
      .miniCard .actionText{min-height:0}
      pre{height:58vh;min-height:420px;font-size:12px;padding:12px}
      .watermark{position:static;text-align:center;padding:12px}
    }
  </style>
</head>
<body>
<demo-app>
<main>
  <header>
    <div class="brandRow">
      <a class="brandLink" href="/demo" title="Back to DeltaSignal demo landing">
      <img class="logo stageReveal" style="--delay:.62s" src="/assets/deltasignal-app-icon.png" alt="DeltaSignal logo">
      <div>
        <div class="eyebrow stageReveal" style="--delay:.32s">DeltaSignal Evidence OS · Research Memory Proof</div>
        <div class="entryName stageReveal" style="--delay:.42s">Second proof after the flagship Apple workflow</div>
        <h1 class="proofTitle">Run The HUT Research-Memory Proof</h1>
        <p class="stageReveal" style="--delay:1.02s">Paste the private demo key, then run the same deployed Cloud Run surfaces used by the command demo. The key stays in this browser session and is sent only as <span class="mono">X-Demo-Key</span>.</p>
      </div>
      </a>
    </div>
    <div class="headerControls stageReveal" style="--delay:.62s">
      <div class="status"><i class="dot"></i><span id="host"></span></div>
      <div class="headerInputGrid">
        <div>
          <label for="key">Private demo API key</label>
          <input id="key" type="password" autocomplete="off" placeholder="DEMO_KEY">
        </div>
        <div>
          <label for="tripcode">TripCode</label>
          <input id="tripcode" value="TF-SUB-9DA70A7F98">
        </div>
        <div>
          <label for="session">Session ID</label>
          <input id="session" value="browser-hut-demo">
        </div>
      </div>
    </div>
  </header>

  <section class="appGrid">
    <div class="panel controlPanel stageReveal" style="--delay:.78s">
      <div class="panelHeader">
        <div><h2>Run HUT Workflow</h2><p>One TripCode becomes a bounded research packet.</p></div>
        <span class="pill">Live route</span>
      </div>
      <div class="topline">
        <div class="miniCard metric"><strong>1</strong><span>TripCode</span></div>
        <div class="miniCard metric"><strong>10</strong><span>River nodes</span></div>
        <div class="miniCard metric"><strong>A2A</strong><span>Artifact</span></div>
      </div>
      <div class="flow">
        <div><b>Discover</b>Agent Card</div>
        <div><b>Invoke</b>HUT</div>
        <div><b>Verify</b>Packet</div>
        <div><b>Remember</b>Monitor</div>
        <div><b>Control</b>Usage</div>
      </div>
      <div>
      <div class="runGrid">
        <button data-action="overview"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8"></circle><path d="M12 8v4l3 2"></path></svg></span>Overview</button>
        <button data-action="article"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 4h10l4 4v12H5z"></path><path d="M15 4v4h4"></path><path d="M8 12h8M8 16h6"></path></svg></span>Article</button>
        <button class="secondary" data-action="health"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 6 9 17l-5-5"></path></svg></span>Health</button>
        <button class="secondary" data-action="card"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><rect x="4" y="5" width="16" height="14" rx="2"></rect><path d="M7 9h10M7 13h6"></path></svg></span>Card</button>
        <button data-action="resolve"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M13 2 4 14h7l-1 8 10-13h-7z"></path></svg></span>Resolve</button>
        <button data-action="a2aResolve"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M7 7h10v10H7z"></path><path d="M12 2v5M12 17v5M2 12h5M17 12h5"></path></svg></span>A2A</button>
        <button data-action="a2aMonitor"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 12a8 8 0 0 1 14-5"></path><path d="M18 3v4h-4"></path><path d="M20 12a8 8 0 0 1-14 5"></path><path d="M6 21v-4h4"></path></svg></span>Monitor</button>
        <button class="secondary" data-action="usage"><span class="btnIcon"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M4 19V5"></path><path d="M8 19v-6"></path><path d="M12 19V9"></path><path d="M16 19v-3"></path><path d="M20 19V7"></path></svg></span>Usage</button>
      </div>
      </div>
      <div class="miniCard actionCard"><b id="actionTitle">What happens next</b><span id="actionText" class="actionText"></span></div>
    </div>
    <div class="panel responsePanel stageReveal" style="--delay:.92s">
      <div class="panelHeader">
        <div><h2>Response Workspace</h2><p>JSON proof from the deployed Go + Lit Cloud Run app.</p></div>
        <div class="responseTools" aria-label="Response navigation controls">
          <button data-scroll="top" type="button">Top</button>
          <button data-scroll="page" type="button">Scroll</button>
          <button data-scroll="end" type="button">End</button>
          <button class="renderButton" data-render="rich" type="button">Render</button>
          <button data-render="json" type="button">JSON</button>
        </div>
      </div>
      <pre id="out"><span class="resp-title">Loading overview.</span></pre>
    </div>
  </section>
  <section class="captionBar stageReveal" style="--delay:2.55s" aria-label="Demo explanation">
    <span id="captionText" class="captionText"></span>
  </section>
  <div class="watermark">© 2026 AITrailblazer · DeltaSignal</div>
</main>
</demo-app>
<script type="module">
  import {LitElement} from 'https://cdn.jsdelivr.net/npm/lit@3/+esm';

  class DeltaSignalDemoApp extends LitElement {
    createRenderRoot() {
      return this;
    }

    connectedCallback() {
      super.connectedCallback();
      queueMicrotask(() => this.bindControls());
    }

    bindControls() {
      if (this.bound) return;
      this.out = this.querySelector('#out');
      this.key = this.querySelector('#key');
      this.tripcode = this.querySelector('#tripcode');
      this.session = this.querySelector('#session');
      this.actionTitle = this.querySelector('#actionTitle');
      this.actionText = this.querySelector('#actionText');
      this.captionText = this.querySelector('#captionText');
      if (!this.out || !this.key || !this.tripcode || !this.session || !this.actionTitle || !this.actionText || !this.captionText) return;
      this.bound = true;
      this.querySelector('#host').textContent = location.host || 'Cloud Run';
      this.autoplayMode = new URLSearchParams(location.search).get('autoplay') === '1';
      for (const button of this.querySelectorAll('button[data-action]')) {
        button.addEventListener('click', () => this.run(button.dataset.action));
      }
      for (const button of this.querySelectorAll('button[data-scroll]')) {
        button.addEventListener('click', () => this.scrollResponse(button.dataset.scroll, button));
      }
      for (const button of this.querySelectorAll('button[data-render]')) {
        button.addEventListener('click', () => this.renderResponse(button.dataset.render, button));
      }
      this.describeAction('overview');
      this.showOverview(null);
      if (this.autoplayMode) {
        setTimeout(() => this.playTimedDemo(), 1200);
      }
    }

    headers(json = false) {
      const value = this.key.value.trim();
      const h = {};
      if (json) h['Content-Type'] = 'application/json';
      if (value) h['X-Demo-Key'] = value;
      return h;
    }

    async show(label, request, button) {
      this.setRunningButton(button);
      const started = Date.now();
      const startedAt = new Date(started).toISOString();
      this.out.innerHTML = this.traceHTML(label, 'in flight', startedAt, '0.00s') + '<span class="resp-status">fetch() request sent. Waiting for response</span><span class="streamCursor"></span>';
      const timer = setInterval(() => {
        const elapsed = this.querySelector('#elapsed');
        if (elapsed) elapsed.textContent = ((Date.now() - started) / 1000).toFixed(2) + 's';
      }, 80);
      try {
        const res = await request();
        const text = await res.text();
        const duration = ((Date.now() - started) / 1000).toFixed(2) + 's';
        const wait = Math.max(0, 900 - (Date.now() - started));
        if (wait) await new Promise(resolve => setTimeout(resolve, wait));
        let body = text;
        let parsed = null;
        let highlighted = '';
        try {
          parsed = JSON.parse(text);
          body = JSON.stringify(parsed, null, 2);
          highlighted = this.colorizeJSON(body);
        } catch {
          highlighted = this.escapeHTML(body);
        }
        clearInterval(timer);
        const statusClass = res.ok ? 'resp-status' : 'resp-error';
        const trace = this.traceHTML(label, 'completed', startedAt, duration, res.status);
        const header = '<span class="resp-title">' + this.escapeHTML(label) + '</span>\n<span class="' + statusClass + '">HTTP ' + res.status + ' · measured network time ' + duration + '</span>\n\n';
        this.lastResponse = {label, status:res.status, ok:res.ok, startedAt, duration, body, parsed, highlighted, trace, header};
        await this.streamHTML(trace, header, highlighted);
      } catch (err) {
        clearInterval(timer);
        this.out.innerHTML = '<span class="resp-title">' + this.escapeHTML(label) + '</span>\n<span class="resp-error">ERROR</span>\n\n' + this.escapeHTML(err && err.message ? err.message : String(err));
      } finally {
        this.setRunningButton(null);
      }
    }

    describeAction(action) {
      const copy = {
        article: {
          title: 'Article Start',
          text: 'We begin with the public Delta Signal Substack article. The visible subtitle contains the ATLAS-7 TripCode, which becomes the stable handle for the HUT River and evidence packet.',
          caption: 'On screen: the response window shows the top of the published Hut 8 Substack article, the visible article URL, and the TripCode line that starts the demo.',
          scrollCaption: 'Now scrolling the article proof panel: the screenshot, live Substack address, and TripCode are all visible evidence that this workflow starts from a real public research object.'
        },
        overview: {
          title: 'Overview',
          text: 'This first step explains the full demo path: start from Substack research, discover the agent surface, resolve the HUT TripCode, inspect evidence boundaries, then monitor the thesis.',
          caption: 'On screen: the left panel contains the controls, while the blue response workspace shows the objective, five visible steps, and three proof cards.',
          scrollCaption: 'Now scrolling the overview panel: the user can see the complete sequence before running any protected request.'
        },
        health: {
          title: 'Health Check',
          text: 'The browser calls the deployed Go service health route. This proves the runtime is reachable before any protected research request is made.',
          caption: 'On screen: the blue response workspace turns into a live request trace, then prints the health JSON returned by the deployed service.'
        },
        card: {
          title: 'Agent Discovery',
          text: 'The browser fetches the A2A Agent Card. A client can inspect skills, input modes, output modes, and provider metadata before choosing a workflow.',
          caption: 'On screen: the JSON response lists the Agent Card fields, callable skills, examples, provider metadata, and output modes.',
          scrollCaption: 'Now scrolling the Agent Card: this is the discoverable A2A surface, including skills for resolving TripCodes, monitoring a thesis, and running Track 3 scenarios.'
        },
        resolve: {
          title: 'Resolve TripCode',
          text: 'The protected Cloud Run route resolves the HUT TripCode into article memory, River context, evidence boundaries, execution trace, and usage metadata.',
          caption: 'On screen: the trace shows a protected request in flight, then the highlighted JSON response reveals mode, TripCode, issuer, sources, summary, usage, and disclosures.',
          scrollCaption: 'Now scrolling the resolved packet: watch for execution trace, agent context sources, research packet fields, Gemini summary, usage accounting, and boundary disclosures.'
        },
        a2aResolve: {
          title: 'A2A Resolve',
          text: 'The UI sends a JSON-RPC A2A message. The agent interprets the user intent, invokes the TripCode research workflow, and returns a structured task artifact.',
          caption: 'On screen: the response is a JSON-RPC task result, showing how another agent would invoke the same HUT TripCode workflow.',
          scrollCaption: 'Now scrolling the A2A task artifact: this shows the agent-facing return shape, not just a browser-only response.'
        },
        render: {
          title: 'Rendered Packet',
          text: 'The latest JSON response is rendered as a readable evidence page: summary first, provenance next, then sources, usage, boundaries, and raw packet details.',
          caption: 'On screen: the same resolve response becomes a rich HTML evidence packet designed for a user to inspect quickly.',
          scrollCaption: 'Now scrolling the rendered packet: summary, provenance, sources, usage accounting, disclosures, and raw evidence remain visible without losing the original JSON shape.'
        },
        a2aMonitor: {
          title: 'Monitor Thesis',
          text: 'The agent uses the same TripCode as a post-publication baseline and returns weakened assumptions, stale evidence checks, invalidation logic, and monitor-next actions.',
          caption: 'On screen: the monitor response is the follow-up state for the same TripCode, with stale-evidence checks and next actions visible in JSON.',
          scrollCaption: 'Now scrolling the thesis monitor result: the same TripCode is reused as a memory baseline for weakened assumptions, invalidation checks, and next evidence actions.'
        },
        usage: {
          title: 'Usage Ledger',
          text: 'The browser calls the usage endpoint so the demo shows cost awareness, request accounting, and remaining Google Cloud credit budget discipline.',
          caption: 'On screen: the usage response shows request accounting and cost fields so the demo remains measurable and budget-aware.',
          scrollCaption: 'Now scrolling usage: the demo keeps request accounting visible so execution remains measurable and cost-aware.'
        }
      }[action] || {title: 'Action', text: 'The selected workflow runs and returns bounded evidence instead of unconstrained model output.'};
      this.actionCopy = this.actionCopy || {};
      this.actionCopy[action] = copy;
      this.actionTitle.textContent = copy.title;
      this.captionText.innerHTML = '';
      this.describeToken = (this.describeToken || 0) + 1;
      const token = this.describeToken;
      return this.revealActionText(copy.text).then(() => {
        if (token === this.describeToken) return this.revealCaption(copy.caption || copy.text);
        return undefined;
      });
    }

    async revealActionText(text) {
      this.actionRevealToken = (this.actionRevealToken || 0) + 1;
      const token = this.actionRevealToken;
      this.actionText.innerHTML = '';
      let shown = '';
      const words = text.split(' ');
      for (const word of words) {
        if (token !== this.actionRevealToken) return;
        shown += (shown ? ' ' : '') + word;
        this.actionText.innerHTML = this.escapeHTML(shown) + '<span class="actionCursor"></span>';
        await new Promise(resolve => setTimeout(resolve, this.autoplayMode ? 52 : 82));
      }
      if (token === this.actionRevealToken) this.actionText.textContent = text;
    }

    async revealCaption(text) {
      this.captionRevealToken = (this.captionRevealToken || 0) + 1;
      const token = this.captionRevealToken;
      this.captionText.innerHTML = '';
      let shown = '';
      const words = text.split(' ');
      for (const word of words) {
        if (token !== this.captionRevealToken) return;
        shown += (shown ? ' ' : '') + word;
        this.captionText.innerHTML = this.escapeHTML(shown) + '<span class="captionCursor"></span>';
        await new Promise(resolve => setTimeout(resolve, this.autoplayMode ? 72 : 105));
      }
      if (token === this.captionRevealToken) this.captionText.textContent = text;
    }

    async autoInspectResponse(action) {
      const copy = this.actionCopy && this.actionCopy[action];
      if (!copy || !copy.scrollCaption || !this.out) return;
      const max = this.out.scrollHeight - this.out.clientHeight;
      if (max < 80) return;
      const token = this.describeToken;
      this.out.scrollTop = 0;
      await new Promise(resolve => setTimeout(resolve, 420));
      if (token !== this.describeToken) return;
      const captionPromise = this.revealCaption(copy.scrollCaption);
      await new Promise(resolve => setTimeout(resolve, 420));
      if (token !== this.describeToken) return;
      await this.animateScroll(0, max, Math.min(this.autoplayMode ? 6800 : 9500, Math.max(this.autoplayMode ? 3200 : 4800, max * (this.autoplayMode ? 5.2 : 7.5))));
      await captionPromise;
    }

    showOverview(button) {
      this.setRunningButton(button);
      this.out.innerHTML = '<span class="overviewStart">' +
        '<span class="objectiveBox">' +
          '<b>Objective</b>' +
          '<span>Show how Delta Signal Gemini AI Agent turns one public Substack TripCode into a bounded HUT issuer intelligence workflow: research memory, River continuity, evidence boundaries, Gemini synthesis, monitor-next actions, and usage visibility.</span>' +
        '</span>' +
        '<span class="traceBox">' +
          '<span class="traceTitle"><i class="traceDot"></i> What to expect</span>' +
          '<span class="traceGrid">' +
            '<b>1</b><span class="route">Article Start: show the public Delta Signal Substack page and TripCode.</span>' +
            '<b>2</b><span>Agent Card: discover the A2A skills before invocation.</span>' +
            '<b>3</b><span>Resolve: call the protected Go Cloud Run route and return the HUT research packet.</span>' +
            '<b>4</b><span>Monitor: reuse the same TripCode as the post-publication thesis baseline.</span>' +
            '<b>5</b><span>Usage: show cost-aware request accounting for the demo session.</span>' +
          '</span>' +
        '</span>' +
        '<span class="overviewGrid">' +
          '<span class="miniCard"><b>Substack starts the loop</b><span>The article is visible to the user and carries the TripCode in the subtitle.</span></span>' +
          '<span class="miniCard"><b>ATLAS-7 is the evidence engine</b><span>Resolved packets preserve source dates, caveats, stale flags, and boundaries where available.</span></span>' +
          '<span class="miniCard"><b>Gemini orchestrates synthesis</b><span>The agent converts intent into a bounded diligence packet instead of generic chat.</span></span>' +
        '</span>' +
      '</span>';
      this.out.scrollTop = 0;
      setTimeout(() => this.setRunningButton(null), 520);
    }

    showArticle(button) {
      this.setRunningButton(button);
      const tripcode = this.escapeHTML(this.tripcode.value.trim() || 'TF-SUB-9DA70A7F98');
      this.out.innerHTML = '<span class="articleStart">' +
        '<span class="objectiveBox">' +
          '<b>Objective</b>' +
          '<span>Show a real HUT research workflow: start from the published Delta Signal Substack article, read the TripCode, then let the deployed Gemini AI Agent resolve the River, evidence boundaries, thesis deltas, monitor-next actions, and usage record.</span>' +
        '</span>' +
        '<span class="traceBox">' +
          '<span class="traceTitle"><i class="traceDot"></i> Starting research object</span>' +
          '<span class="traceGrid">' +
            '<b>source</b><span class="route">Delta Signal Substack article</span>' +
            '<b>article</b><span>Hut 8: The Re-Rating Has A Deadline</span>' +
            '<b>substack</b><span class="route">https://deltasignal.substack.com/p/hut-8-the-re-rating-has-a-deadline</span>' +
            '<b>tripcode</b><span class="ok">' + tripcode + '</span>' +
            '<b>role</b><span>Public article starts the River; the agent resolves the code into evidence memory.</span>' +
          '</span>' +
        '</span>' +
        '<span class="articleHero">' +
          '<img src="/assets/substack-hut-article-top.png" alt="Top of the Hut 8 Substack article showing the TripCode subtitle">' +
          '<span class="substackURL">https://deltasignal.substack.com/p/hut-8-the-re-rating-has-a-deadline</span>' +
          '<a class="articleLink" href="https://deltasignal.substack.com/p/hut-8-the-re-rating-has-a-deadline" target="_blank" rel="noreferrer">Open live Substack article</a>' +
        '</span>' +
        '<span class="articleMeta">' +
          '<span class="miniCard"><b>What the user sees first</b><span>A published Substack research article with a subtitle containing the ATLAS-7 TripCode.</span></span>' +
          '<span class="miniCard"><b>What the agent uses next</b><span><code>' + tripcode + '</code> becomes the resolver key for River memory, evidence boundaries, and HUT follow-up tasks.</span></span>' +
        '</span>' +
      '</span>';
      this.out.scrollTop = 0;
      setTimeout(() => this.setRunningButton(null), 520);
    }

    async playTimedDemo() {
      const sequence = [
        ['article', 8500],
        ['card', 8500],
        ['resolve', 14500],
        ['render', 9500],
        ['a2aResolve', 13000],
        ['a2aMonitor', 11500],
        ['usage', 7500],
      ];
      for (const [action, pause] of sequence) {
        const button = this.querySelector('button[data-action="' + action + '"]');
        if (button) {
          button.scrollIntoView({block: 'nearest', inline: 'nearest'});
          button.classList.add('running');
          setTimeout(() => button.classList.remove('running'), 950);
        }
        await this.run(action);
        await new Promise(resolve => setTimeout(resolve, pause));
      }
      this.revealCaption('Demo complete: one browser flow started from Substack, discovered the agent surface, resolved the HUT TripCode, returned evidence-bounded outputs, monitored the thesis, and showed usage visibility.');
    }

    setRunningButton(button) {
      for (const item of this.querySelectorAll('button.running')) item.classList.remove('running');
      if (button) button.classList.add('running');
    }

    async streamHTML(trace, header, highlighted) {
      this.out.innerHTML = trace + header + '<span id="streamTarget"></span><span class="streamCursor"></span>';
      const target = this.querySelector('#streamTarget');
      const chunks = this.chunkHighlightedHTML(highlighted, 95);
      let rendered = '';
      for (const chunk of chunks) {
        rendered += chunk;
        target.innerHTML = rendered;
        this.out.scrollTop = 0;
        await new Promise(resolve => setTimeout(resolve, 18));
      }
      this.out.innerHTML = trace + header + highlighted;
      this.out.scrollTop = 0;
    }

    scrollResponse(mode, button) {
      for (const item of this.querySelectorAll('button[data-scroll].active')) item.classList.remove('active');
      button.classList.add('active');
      const start = this.out.scrollTop;
      let target = 0;
      if (mode === 'page') target = Math.min(this.out.scrollHeight - this.out.clientHeight, start + Math.round(this.out.clientHeight * 0.78));
      if (mode === 'end') target = this.out.scrollHeight - this.out.clientHeight;
      this.animateScroll(start, Math.max(0, target), 950).finally(() => button.classList.remove('active'));
    }

    renderResponse(mode, button) {
      for (const item of this.querySelectorAll('button[data-render].active')) item.classList.remove('active');
      if (button) button.classList.add('active');
      if (mode === 'json') {
        if (!this.lastResponse) {
          this.revealCaption('Run Health, Card, Resolve, A2A, Monitor, or Usage first. Then JSON can restore the raw highlighted response.');
          if (button) setTimeout(() => button.classList.remove('active'), 900);
          return;
        }
        this.out.innerHTML = this.lastResponse.trace + this.lastResponse.header + this.lastResponse.highlighted;
        this.out.scrollTop = 0;
        this.revealCaption('Raw JSON view restored: this is the exact response shape returned by the deployed route.');
        if (button) setTimeout(() => button.classList.remove('active'), 900);
        return;
      }
      if (!this.lastResponse || !this.lastResponse.parsed) {
        this.revealCaption('Run a JSON-producing step first, then click Render to turn the latest response into a rich HTML evidence page.');
        if (button) setTimeout(() => button.classList.remove('active'), 900);
        return;
      }
      this.out.innerHTML = this.richResponseHTML(this.lastResponse);
      this.out.scrollTop = 0;
      this.revealCaption('Rendered view: the same JSON response is now reorganized as a readable evidence page with summary, provenance, sources, usage, boundaries, and raw packet sections.');
      if (button) setTimeout(() => button.classList.remove('active'), 900);
    }

    async showRenderedPacket() {
      if (!this.lastResponse || !this.lastResponse.parsed) {
        this.out.innerHTML = '<span class="resp-title">Rendered Packet</span>\n<span class="resp-error">No JSON response is available yet.</span>\n\nRun Resolve first, then render the evidence packet.';
        this.out.scrollTop = 0;
        return;
      }
      this.out.innerHTML = this.richResponseHTML(this.lastResponse);
      this.out.scrollTop = 0;
      await new Promise(resolve => setTimeout(resolve, 420));
    }

    richResponseHTML(response) {
      const data = response.parsed || {};
      const pick = (...paths) => {
        for (const path of paths) {
          let value = data;
          for (const key of path.split('.')) {
            if (value == null || typeof value !== 'object' || !(key in value)) {
              value = undefined;
              break;
            }
            value = value[key];
          }
          if (value !== undefined && value !== null && value !== '') return value;
        }
        return '';
      };
      const label = this.escapeHTML(response.label || 'DeltaSignal response');
      const mode = this.escapeHTML(String(pick('mode', 'result.mode', 'artifact.mode') || 'response'));
      const tripcode = this.escapeHTML(String(pick('tripcode', 'result.tripcode', 'artifact.tripcode', 'research_packet.article.tripcode') || this.tripcode.value.trim() || 'TF-SUB-9DA70A7F98'));
      const issuer = this.escapeHTML(String(pick('issuer', 'result.issuer', 'artifact.issuer', 'research_packet.article.primary_issuer') || 'HUT'));
      const summary = this.escapeHTML(this.readableText(pick('gemini_summary', 'result.gemini_summary', 'artifact.gemini_summary', 'research_packet.thesis_map.summary')) || 'Structured response returned by the DeltaSignal Gemini AI Agent.');
      const execution = pick('execution_trace', 'result.execution_trace', 'artifact.execution_trace');
      const sources = pick('agent_context.sources', 'result.agent_context.sources', 'artifact.agent_context.sources', 'research_packet.sources');
      const usage = pick('usage', 'cost', 'result.usage', 'result.cost', 'artifact.usage');
      const boundaries = pick('boundaries', 'disclosures', 'result.boundaries', 'result.disclosures', 'artifact.boundaries', 'research_packet.boundaries');
      return '<span class="richRender">' +
        '<span class="richHero">' +
          '<span class="richKicker">Rendered evidence page · same response, human-readable surface</span>' +
          '<span class="richTitle">' + label + '</span>' +
          '<span class="richSummary">' + summary + '</span>' +
        '</span>' +
        '<span class="richGrid">' +
          this.richMetric('TripCode', '<code>' + tripcode + '</code>', 'gold') +
          this.richMetric('Issuer', issuer, 'green') +
          this.richMetric('HTTP / Time', 'HTTP ' + this.escapeHTML(response.status) + ' · ' + this.escapeHTML(response.duration), '') +
        '</span>' +
        '<span class="richSection"><h3>Workflow Identity</h3>' +
          '<span class="richTable">' +
            this.richRow('mode', mode) +
            this.richRow('tripcode', tripcode) +
            this.richRow('issuer', issuer) +
            this.richRow('started', this.escapeHTML(response.startedAt)) +
          '</span>' +
        '</span>' +
        this.richObjectSection('Execution Trace', execution) +
        this.richSourcesSection(sources) +
        this.richObjectSection('Usage / Cost Controls', usage) +
        this.richObjectSection('Boundaries / Disclosures', boundaries) +
        this.richObjectSection('Research Packet / Artifact', pick('research_packet', 'artifact', 'result.artifact', 'result')) +
        '<span class="richSection"><h3>Raw Response</h3><span class="richPre">' + this.escapeHTML(response.body) + '</span></span>' +
      '</span>';
    }

    richMetric(title, value, tone) {
      return '<span class="richCard ' + this.escapeHTML(tone || '') + '"><b>' + this.escapeHTML(title) + '</b><span>' + value + '</span></span>';
    }

    richRow(key, value) {
      return '<span class="richRow"><b>' + this.escapeHTML(key) + '</b><span>' + value + '</span></span>';
    }

    richObjectSection(title, value) {
      if (value === undefined || value === null || value === '') {
        return '<span class="richSection"><h3>' + this.escapeHTML(title) + '</h3><p>No field was returned in this response. The absence stays explicit instead of being inferred.</p></span>';
      }
      if (Array.isArray(value)) {
        return '<span class="richSection"><h3>' + this.escapeHTML(title) + '</h3><span class="richGrid">' + value.slice(0, 6).map((item, index) => this.richMetric(String(index + 1), this.escapeHTML(this.readableText(item)), '')).join('') + '</span></span>';
      }
      if (typeof value === 'object') {
        const rows = Object.entries(value).slice(0, 12).map(([key, item]) => this.richRow(key, this.escapeHTML(this.readableText(item)))).join('');
        return '<span class="richSection"><h3>' + this.escapeHTML(title) + '</h3><span class="richTable">' + rows + '</span></span>';
      }
      return '<span class="richSection"><h3>' + this.escapeHTML(title) + '</h3><p>' + this.escapeHTML(String(value)) + '</p></span>';
    }

    richSourcesSection(sources) {
      if (!Array.isArray(sources) || sources.length === 0) {
        return '<span class="richSection"><h3>Agent Context Sources</h3><p>No source list was returned in this response. Missing source metadata remains visible.</p></span>';
      }
      return '<span class="richSection"><h3>Agent Context Sources</h3><span class="richGrid">' + sources.slice(0, 6).map((source, index) => {
        const url = this.escapeHTML(String(source.url || source.href || source.source || 'source-' + (index + 1)));
        const status = this.escapeHTML(String(source.status || source.http_status || source.state || 'returned'));
        const hash = this.escapeHTML(String(source.sha256 || source.hash || source.content_sha256 || 'hash not returned'));
        return '<span class="richCard"><b>Source ' + (index + 1) + '</b><span><code>' + url + '</code><br>Status: ' + status + '<br>Hash: ' + hash + '</span></span>';
      }).join('') + '</span></span>';
    }

    readableText(value) {
      if (value === undefined || value === null) return '';
      if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') return String(value);
      if (Array.isArray(value)) return value.map(item => this.readableText(item)).filter(Boolean).join(' · ');
      if (typeof value === 'object') {
        return Object.entries(value).slice(0, 8).map(([key, item]) => key + ': ' + this.readableText(item)).join(' · ');
      }
      return String(value);
    }

    async animateScroll(start, target, duration) {
      const startTime = Date.now();
      while (true) {
        const t = Math.min(1, (Date.now() - startTime) / duration);
        const eased = t < .5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2;
        this.out.scrollTop = start + (target - start) * eased;
        if (t >= 1) break;
        await new Promise(resolve => setTimeout(resolve, 16));
      }
    }

    chunkHighlightedHTML(html, visibleChars) {
      const chunks = [];
      let current = '';
      let visible = 0;
      for (let i = 0; i < html.length;) {
        if (html[i] === '<') {
          const end = html.indexOf('>', i);
          if (end === -1) break;
          current += html.slice(i, end + 1);
          i = end + 1;
          continue;
        }
        current += html[i];
        visible++;
        i++;
        if (visible >= visibleChars || html[i - 1] === '\n') {
          chunks.push(current);
          current = '';
          visible = 0;
        }
      }
      if (current) chunks.push(current);
      return chunks;
    }

    traceHTML(label, state, startedAt, elapsed, status) {
      const stateClass = state === 'completed' ? 'ok' : '';
      return '<span class="traceBox">' +
        '<span class="traceTitle"><i class="traceDot"></i> Live request trace</span>' +
        '<span class="traceGrid">' +
        '<b>route</b><span class="route">' + this.escapeHTML(label) + '</span>' +
        '<b>state</b><span class="' + stateClass + '">' + this.escapeHTML(state) + '</span>' +
        '<b>started</b><span>' + this.escapeHTML(startedAt) + '</span>' +
        '<b>elapsed</b><span id="elapsed">' + this.escapeHTML(elapsed) + '</span>' +
        (status ? '<b>status</b><span class="ok">HTTP ' + this.escapeHTML(status) + '</span>' : '') +
        '</span></span>';
    }

    escapeHTML(value) {
      return String(value).replace(/[&<>"']/g, function(ch) {
        return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[ch];
      });
    }

    colorizeJSON(json) {
      let out = '';
      let i = 0;
      while (i < json.length) {
        const ch = json[i];
        if (ch === '"') {
          const start = i;
          i++;
          let escaped = false;
          while (i < json.length) {
            const c = json[i];
            if (escaped) {
              escaped = false;
            } else if (c === '\\') {
              escaped = true;
            } else if (c === '"') {
              i++;
              break;
            }
            i++;
          }
          let j = i;
          while (j < json.length && /\s/.test(json[j])) j++;
          const cls = json[j] === ':' ? 'json-key' : 'json-string';
          out += '<span class="' + cls + '">' + this.escapeHTML(json.slice(start, i)) + '</span>';
          continue;
        }
        if (ch === '-' || (ch >= '0' && ch <= '9')) {
          const match = json.slice(i).match(/^-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?/);
          if (match) {
            out += '<span class="json-number">' + match[0] + '</span>';
            i += match[0].length;
            continue;
          }
        }
        if (json.startsWith('true', i) || json.startsWith('false', i)) {
          const value = json.startsWith('true', i) ? 'true' : 'false';
          out += '<span class="json-bool">' + value + '</span>';
          i += value.length;
          continue;
        }
        if (json.startsWith('null', i)) {
          out += '<span class="json-null">null</span>';
          i += 4;
          continue;
        }
        if (ch === '{' || ch === '}' || ch === '[' || ch === ']') {
          out += '<span class="json-brace">' + ch + '</span>';
          i++;
          continue;
        }
        if (ch === ':') {
          out += '<span class="json-colon">:</span>';
          i++;
          continue;
        }
        if (ch === ',') {
          out += '<span class="json-punc">,</span>';
          i++;
          continue;
        }
        out += this.escapeHTML(ch);
        i++;
      }
      return out;
    }

    async run(action) {
      const actions = {
        overview: button => this.showOverview(button),
        article: button => this.showArticle(button),
        health: button => this.show('GET /health', () => fetch('/health'), button),
        card: button => this.show('GET /.well-known/agent-card.json', () => fetch('/.well-known/agent-card.json'), button),
        resolve: button => this.show('GET /resolve HUT TripCode', () => fetch('/resolve?tripcode=' + encodeURIComponent(this.tripcode.value.trim()) + '&session_id=' + encodeURIComponent(this.session.value.trim()) + '&issuer=HUT&payload_mode=compact&include_filing_evidence=true&include_prior_articles=true&include_thesis_map=true&include_agent_context=true', {headers: this.headers()}), button),
        a2aResolve: button => this.show('POST /a2a resolve', () => fetch('/a2a', {method:'POST', headers: this.headers(true), body: JSON.stringify({jsonrpc:'2.0', id:'browser-resolve', method:'message/send', params:{message:{parts:[{kind:'text', type:'text/plain', text:'Resolve ' + this.tripcode.value.trim() + ' and show what changed across the HUT River.'}]}, metadata:{session_id:this.session.value.trim()}}})}), button),
        a2aMonitor: button => this.show('POST /a2a monitor', () => fetch('/a2a', {method:'POST', headers: this.headers(true), body: JSON.stringify({jsonrpc:'2.0', id:'browser-monitor', method:'monitor_tripcode_thesis', params:{message:{parts:[{kind:'text', type:'text/plain', text:'Monitor ' + this.tripcode.value.trim() + ' for weakened assumptions, stale evidence, invalidation checks, and monitor-next actions.'}]}, metadata:{session_id:this.session.value.trim()}}})}), button),
        usage: button => this.show('GET /v1/usage', () => fetch('/v1/usage', {headers: this.headers()}), button),
        render: () => this.showRenderedPacket(),
      };
      const button = this.querySelector('button[data-action="' + action + '"]');
      const describePromise = this.describeAction(action);
      const result = await actions[action](button);
      await describePromise;
      await this.autoInspectResponse(action);
      return result;
    }
  }

  customElements.define('demo-app', DeltaSignalDemoApp);
</script>
</body>
</html>`

const rootLandingHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="description" content="DeltaSignal is a governed evidence operating system for financial agents, demonstrated first through an inspectable Apple reference workflow.">
  <title>DeltaSignal Evidence OS · Apple Reference First</title>
  <style>
    :root{color-scheme:dark;--bg:#05070b;--panel:#101722;--line:rgba(255,255,255,.14);--text:#f7fbff;--muted:#b9c7d8;--green:#5df58b;--blue:#72d9ff;--gold:#d6ad5c}
    *{box-sizing:border-box}
    html,body{margin:0;min-height:100%;background:#05070b;color:var(--text);font:16px/1.45 Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;-webkit-font-smoothing:antialiased;text-rendering:geometricPrecision}
    body{display:grid;place-items:center;min-height:100vh;padding:24px;background:radial-gradient(circle at 18% 6%,rgba(93,245,139,.19),transparent 34%),radial-gradient(circle at 84% 10%,rgba(114,217,255,.16),transparent 34%),linear-gradient(135deg,#04070c 0%,#08111a 52%,#07120d 100%)}
    main{width:min(1060px,100%);display:grid;grid-template-columns:1.02fr .98fr;gap:18px;align-items:stretch}
    .panel{border:1px solid rgba(214,173,92,.42);border-radius:26px;background:linear-gradient(180deg,rgba(18,28,43,.96),rgba(7,12,20,.96));box-shadow:0 28px 90px rgba(0,0,0,.42),inset 0 1px 0 rgba(255,255,255,.08);padding:30px;overflow:hidden}
    .brand{display:flex;align-items:center;gap:14px;margin-bottom:28px}
    .logo{width:58px;height:58px;border-radius:16px;object-fit:cover;box-shadow:0 0 0 1px rgba(255,255,255,.18),0 0 34px rgba(214,173,92,.34)}
    .eyebrow{color:var(--gold);text-transform:uppercase;letter-spacing:.12em;font-weight:950;font-size:12px}
    .entry{font-weight:950;text-transform:uppercase;letter-spacing:.02em;margin-top:4px}
    h1{font-size:clamp(42px,6vw,76px);line-height:.92;margin:0 0 18px;letter-spacing:0}
    p{margin:0;color:var(--muted);font-size:19px;max-width:680px}
    .buttons{display:flex;gap:12px;align-items:center;flex-wrap:wrap;margin-top:30px}
    .button{display:inline-flex;align-items:center;gap:11px;border:1px solid rgba(93,245,139,.52);border-radius:999px;padding:16px 20px;background:linear-gradient(135deg,rgba(93,245,139,.92),rgba(114,217,255,.72));color:#051018;text-decoration:none;font-weight:1000;font-size:19px;box-shadow:0 18px 56px rgba(93,245,139,.24);transition:transform .16s ease,box-shadow .16s ease}
    .button:hover{transform:translateY(-2px);box-shadow:0 24px 70px rgba(93,245,139,.34)}
    .button.secondary{border-color:rgba(214,173,92,.58);background:rgba(214,173,92,.1);color:#ffe2a3;box-shadow:none}
    .button svg{width:22px;height:22px;stroke:currentColor;stroke-width:2.5;fill:none;stroke-linecap:round;stroke-linejoin:round}
    .visual{display:grid;gap:12px}
    .chip{display:inline-flex;width:max-content;align-items:center;gap:8px;border:1px solid rgba(93,245,139,.35);border-radius:999px;padding:10px 13px;color:var(--green);font-weight:950;background:rgba(93,245,139,.08)}
    .dot{width:9px;height:9px;border-radius:999px;background:var(--green);box-shadow:0 0 18px rgba(93,245,139,.75)}
    .flow{margin-top:16px;border:1px solid rgba(114,217,255,.2);border-radius:22px;background:radial-gradient(circle at 52% 40%,rgba(93,245,139,.18),transparent 42%),rgba(2,14,21,.68);padding:22px;display:grid;gap:14px}
    .node{border:1px solid rgba(255,255,255,.12);border-radius:16px;padding:15px;background:rgba(255,255,255,.045)}
    .node b{display:block;font-size:18px;color:#fff}
    .node span{display:block;margin-top:4px;color:var(--muted);font-weight:750}
    footer{margin-top:24px;color:rgba(247,251,255,.48);font-size:12px;text-transform:uppercase;letter-spacing:.12em;font-weight:900}
    @media(max-width:820px){body{display:block}main{grid-template-columns:1fr}.panel{padding:22px}h1{font-size:46px}}
  </style>
</head>
<body>
<main>
  <section class="panel">
    <div class="brand">
      <img class="logo" src="/assets/deltasignal-app-icon.png" alt="DeltaSignal logo">
      <div>
        <div class="eyebrow">DeltaSignal Evidence OS · Governed Agent Infrastructure</div>
        <div class="entry">Apple reference first · reusable across issuers</div>
      </div>
    </div>
    <h1>Evidence agents can inspect—not just repeat.</h1>
    <p>Start with the flagship Apple workflow: dated SEC/XBRL facts move through ATLAS-7 interpretation, bounded specialist review, and a provenance-rich client packet. Then inspect HUT as the second proof for continuing research memory.</p>
    <div class="buttons">
      <a class="button" href="https://aitrailblazer.github.io/deltasignal-ai-agent/client-demo/google-cloud-aapl/"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h14"></path><path d="M13 6l6 6-6 6"></path></svg>Open Apple Reference</a>
      <a class="button secondary" href="/demo"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z"></path></svg>Explore HUT Memory Proof</a>
    </div>
    <footer>© 2026 AITrailblazer · DeltaSignal</footer>
  </section>
  <section class="panel visual">
    <div class="chip"><span class="dot"></span> SEC/XBRL · ATLAS-7 · MCP · Google Cloud Agents</div>
    <div class="flow" aria-label="DeltaSignal execution loop">
      <div class="node"><b>① Source · Apple SEC/XBRL</b><span>Dated filing evidence, identity, period, and missingness remain attached.</span></div>
      <div class="node"><b>② Govern · ATLAS-7 + MCP</b><span>Read-only contracts separate facts, calculations, interpretation, and limits.</span></div>
      <div class="node"><b>③ Review · Specialist Agents</b><span>Bounded roles review quality, provenance, applicability, and unresolved claims.</span></div>
      <div class="node"><b>④ Deliver · Evidence Packet</b><span>Clients receive facts, labeled interpretation, confidence, caveats, and reusable follow-ups.</span></div>
    </div>
  </section>
</main>
</body>
</html>`

const demoLandingHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta name="description" content="Explore DeltaSignal's HUT research-memory proof, the second reusable workflow after the flagship Apple evidence reference.">
  <title>DeltaSignal · HUT Research Memory Proof</title>
  <style>
    :root{color-scheme:dark;--bg:#05070b;--panel:#101722;--line:rgba(255,255,255,.14);--text:#f7fbff;--muted:#b9c7d8;--green:#5df58b;--blue:#72d9ff;--gold:#d6ad5c;--amber:#f8d998;--mono:"SF Mono","JetBrains Mono","Cascadia Code",ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}
    *{box-sizing:border-box}
    html,body{margin:0;min-height:100%;background:#05070b;color:var(--text);font:16px/1.42 Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;-webkit-font-smoothing:antialiased;text-rendering:geometricPrecision}
    body{overflow:hidden;background:radial-gradient(circle at 12% 0,rgba(93,245,139,.18),transparent 30%),radial-gradient(circle at 88% 4%,rgba(114,217,255,.18),transparent 34%),linear-gradient(135deg,#04070c 0%,#08111a 48%,#07120d 100%)}
    body::before{content:"";position:fixed;inset:0;pointer-events:none;background-image:linear-gradient(rgba(255,255,255,.028) 1px,transparent 1px),linear-gradient(90deg,rgba(255,255,255,.028) 1px,transparent 1px);background-size:46px 46px;mask-image:linear-gradient(to bottom,rgba(0,0,0,.82),transparent 76%)}
    main{height:100vh;max-width:1320px;margin:0 auto;padding:16px;display:grid;grid-template-rows:auto minmax(0,1fr);gap:14px;position:relative;overflow:hidden}
    header,.panel,.card{border:1px solid var(--line);background:linear-gradient(180deg,rgba(18,28,43,.98),rgba(7,12,20,.98));border-radius:18px;box-shadow:0 22px 64px rgba(0,0,0,.38),inset 0 1px 0 rgba(255,255,255,.08)}
    header{display:flex;align-items:center;justify-content:space-between;gap:16px;padding:14px 16px;border-color:rgba(214,173,92,.42)}
    .brand{display:flex;align-items:center;gap:13px;min-width:0}
    .logo{width:50px;height:50px;border-radius:14px;object-fit:cover;box-shadow:0 0 0 1px rgba(255,255,255,.18),0 0 30px rgba(214,173,92,.34);flex:0 0 auto}
    .eyebrow{color:var(--gold);text-transform:uppercase;letter-spacing:.12em;font-weight:950;font-size:12px}
    .entry{font-weight:950;text-transform:uppercase;letter-spacing:.02em;margin-top:3px}
    .status{display:inline-flex;align-items:center;gap:8px;border:1px solid rgba(93,245,139,.35);border-radius:999px;padding:9px 12px;color:var(--green);font-weight:950;background:rgba(93,245,139,.08)}
    .dot{width:9px;height:9px;border-radius:999px;background:var(--green);box-shadow:0 0 18px rgba(93,245,139,.75)}
    .hero{display:grid;grid-template-columns:38.2fr 61.8fr;gap:14px;min-height:0}
    .panel{padding:18px;min-height:0;overflow:hidden}
    h1{font-size:clamp(34px,3.85vw,52px);line-height:.98;margin:8px 0 11px;letter-spacing:0;max-width:760px}
    h2{font-size:24px;line-height:1.08;margin:0;color:var(--green)}
    p{margin:0;color:var(--muted)}
    .lead{font-size:19px;line-height:1.3;color:#f6fbff;font-weight:780;max-width:720px}
    .sub{font-size:14px;line-height:1.42;margin-top:12px;max-width:700px}
    .startCallout{display:block;margin-top:10px;border:2px solid rgba(248,217,152,.56);border-radius:14px;padding:8px 10px;color:#d3a63a;background:linear-gradient(135deg,#07090e,#10141c 52%,#22190b);box-shadow:0 0 0 1px rgba(0,0,0,.72),0 0 26px rgba(214,173,92,.12),inset 0 1px 0 rgba(255,255,255,.10);font:950 14px/1.08 "Arial Narrow","Roboto Condensed","Aptos Narrow","Inter Tight",Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;letter-spacing:.01em;text-shadow:0 2px 0 rgba(0,0,0,.74);max-height:46px;overflow:hidden}
    .goRow{display:flex;gap:10px;align-items:center;flex-wrap:wrap;margin-top:14px}
    .goButton{display:inline-flex;align-items:center;gap:10px;border:1px solid rgba(93,245,139,.52);border-radius:999px;padding:14px 18px;background:linear-gradient(135deg,rgba(93,245,139,.92),rgba(114,217,255,.72));color:#051018;text-decoration:none;font-weight:1000;font-size:18px;box-shadow:0 16px 48px rgba(93,245,139,.22);transition:transform .16s ease,box-shadow .16s ease}
    .goButton svg{width:21px;height:21px;stroke:currentColor;stroke-width:2.5;fill:none;stroke-linecap:round;stroke-linejoin:round;flex:0 0 auto}
    .goButton:hover{transform:translateY(-2px);box-shadow:0 20px 64px rgba(93,245,139,.32)}
    .demoButton{border-color:rgba(214,173,92,.64);background:linear-gradient(135deg,rgba(214,173,92,.95),rgba(255,236,181,.82));box-shadow:0 16px 48px rgba(214,173,92,.22)}
    .pill{display:inline-flex;align-items:center;border:1px solid rgba(214,173,92,.36);border-radius:999px;padding:9px 11px;color:#ffe2a3;background:rgba(214,173,92,.08);font-weight:950;font-size:12px;text-transform:uppercase;letter-spacing:.08em}
    .proofGrid{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:8px;margin-top:10px}
    .card{padding:10px;border-radius:14px;background:linear-gradient(180deg,rgba(255,255,255,.065),rgba(255,255,255,.026))}
    .card strong{display:block;color:var(--green);font-size:21px;line-height:1}
    .card span{display:block;color:var(--muted);font-size:11px;font-weight:850;margin-top:5px}
    .screenFrame{height:100%;min-height:0;position:relative;border:1px solid rgba(114,217,255,.22);border-radius:18px;background:linear-gradient(180deg,#111928,#07101d);overflow:hidden;box-shadow:0 20px 70px rgba(0,0,0,.42),0 0 48px rgba(114,217,255,.10)}
    .screenFrame::before{content:"";position:absolute;left:0;right:0;top:0;height:34px;background:rgba(255,255,255,.06);border-bottom:1px solid rgba(255,255,255,.1)}
    .screenDots{position:absolute;top:12px;left:14px;display:flex;gap:7px;z-index:2}
    .screenDots span{width:9px;height:9px;border-radius:999px;background:#ff5f57}
    .screenDots span:nth-child(2){background:#ffbd2e}.screenDots span:nth-child(3){background:#28c840}
    .architectureStage{position:absolute;inset:44px 14px 14px;border:1px solid rgba(255,255,255,.08);border-radius:16px;background:radial-gradient(circle at 52% 42%,rgba(93,245,139,.17),transparent 34%),radial-gradient(circle at 78% 20%,rgba(114,217,255,.11),transparent 26%),rgba(2,14,21,.66);overflow:hidden}
    .archSvg{width:100%;height:100%;display:block}
    .archNode{fill:#121b2b;stroke:#4b95bd;stroke-width:3}
    .archNode.primary{stroke:#5df58b;filter:drop-shadow(0 0 18px rgba(93,245,139,.34))}
    .archIcon{fill:none;stroke:#dff6ff;stroke-width:1.7;stroke-linecap:round;stroke-linejoin:round;opacity:.48}
    .archIcon.hot{stroke:#5df58b;filter:drop-shadow(0 0 3px rgba(93,245,139,.14))}
    .archIcon.blue{stroke:#72d9ff;filter:drop-shadow(0 0 3px rgba(114,217,255,.13))}
    .archLabel{fill:#f7fbff;font-size:24px;font-weight:900}
    .archSub{fill:#bcc9dc;font-size:18px;font-weight:820}
    .archWire{fill:none;stroke:#72d9ff;stroke-width:4;stroke-dasharray:14 14;animation:archDash 1.4s linear infinite}
    .archWire.green{stroke:#5df58b;filter:drop-shadow(0 0 10px rgba(93,245,139,.28))}
    .archPulse{opacity:.16;fill:#5df58b;animation:archPulse 2.2s ease-in-out infinite;transform-origin:center}
    @keyframes archDash{to{stroke-dashoffset:-56}}
    @keyframes archPulse{0%,100%{opacity:.10;transform:scale(.94)}48%{opacity:.28;transform:scale(1.06)}}
    .screenInner{position:absolute;inset:44px 14px 14px;display:grid;grid-template-columns:.92fr 1.08fr;gap:12px}
    .screenHero{border:1px solid rgba(93,245,139,.25);border-radius:14px;background:radial-gradient(circle at 18% 18%,rgba(93,245,139,.15),transparent 42%),#091120;padding:15px}
    .screenHeroHead{display:flex;align-items:center;gap:8px}
    .screenAppIcon{width:34px;height:34px;border-radius:10px;object-fit:cover;box-shadow:0 0 18px rgba(255,124,47,.22)}
    .screenHero .tiny{display:inline-block;color:#f8d998;font-size:9px;font-weight:950;letter-spacing:.1em;text-transform:uppercase;border:1px solid rgba(248,217,152,.28);border-radius:999px;padding:4px 6px}
    .screenHero strong{display:block;margin-top:12px;font-size:25px;line-height:1.02;color:#fff8ef}
    .screenHero p{font-size:18px;line-height:1.22;color:#dce7f6;margin:11px 0 0;font-weight:820}
    .pipelineShot{width:100%;height:170px;object-fit:cover;object-position:center;border-radius:13px;border:1px solid rgba(114,217,255,.22);margin-top:13px;box-shadow:0 18px 42px rgba(0,0,0,.34)}
    .screenLoop{display:grid;gap:8px}
    .screenNode{display:grid;grid-template-columns:32px 1fr;gap:9px;align-items:center;border:1px solid rgba(255,255,255,.1);border-radius:11px;background:rgba(255,255,255,.045);padding:9px;animation:screenNodePulse 2.6s ease-in-out infinite;animation-delay:calc(var(--i,0) * .18s)}
    .screenNode i{display:grid;place-items:center;width:30px;height:30px;border-radius:999px;background:#5df58b;color:#061018;font-style:normal;font-weight:950}
    .screenNode b{display:block;color:#f7fbff;font-size:13px;line-height:1.1}
    .screenNode span{display:block;color:#aebbd0;font-size:11px;line-height:1.2;margin-top:3px}
    @keyframes screenNodePulse{0%,100%{transform:translateX(0);border-color:rgba(255,255,255,.1)}45%{transform:translateX(5px);border-color:rgba(93,245,139,.42)}}
    .footerLine{display:none}
    .landingWatermark{position:fixed;right:22px;bottom:18px;z-index:20;color:rgba(247,251,255,.44);font:850 10px/1.2 Inter,ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;letter-spacing:.09em;text-transform:uppercase;text-shadow:0 1px 12px rgba(0,0,0,.72);pointer-events:none}
    .reveal{opacity:0;transform:translateY(16px);filter:blur(5px);animation:landingReveal .95s cubic-bezier(.2,.72,.18,1) forwards;animation-delay:var(--delay,0s)}
    .screenNode.reveal{animation:landingReveal .95s cubic-bezier(.2,.72,.18,1) forwards, screenNodePulse 2.6s ease-in-out infinite;animation-delay:var(--delay,0s),calc(var(--i,0) * .18s + 2.25s)}
    @keyframes landingReveal{0%{opacity:0;transform:translateY(18px) scale(.985);filter:blur(7px)}70%{opacity:1;filter:blur(0)}100%{opacity:1;transform:translateY(0) scale(1);filter:blur(0)}}
    @media(max-height:760px){
      main{padding:12px;gap:10px}
      header{padding:11px 13px}
      .logo{width:42px;height:42px}
      h1{font-size:clamp(28px,3.55vw,42px)}
      .lead{font-size:17px}
      .sub{font-size:13px;margin-top:10px}
      .proofGrid{margin-top:8px}
      .card{padding:8px}
      .card strong{font-size:18px}
      .card span{font-size:10px}
      .goRow{margin-top:14px}
      .pipelineShot{height:136px}
      .startCallout{font-size:13px;padding:7px 9px;max-height:44px}
    }
    @media(max-width:920px){body{overflow:auto}main{height:auto}.hero{grid-template-columns:1fr}.screenFrame{height:560px}.proofGrid{grid-template-columns:1fr 1fr}header{align-items:flex-start;flex-direction:column}}
  </style>
</head>
<body>
<main>
  <header class="reveal" style="--delay:.08s">
    <div class="brand">
      <img class="logo" src="/assets/deltasignal-app-icon.png" alt="DeltaSignal logo">
      <div>
        <div class="eyebrow">DeltaSignal Evidence OS · Second Product Proof</div>
        <div class="entry">Research memory beyond the flagship Apple reference</div>
      </div>
    </div>
    <div class="status"><span class="dot"></span> Cloud Run · Lit · A2A · TripCode</div>
  </header>
  <section class="hero">
    <div class="panel">
      <span class="pill reveal" style="--delay:.20s">Second proof · continuing research memory</span>
      <h1 class="reveal" style="--delay:.30s">A published thesis becomes an inspectable agent workflow.</h1>
      <p class="lead reveal" style="--delay:.42s">After Apple proves the filing-to-packet path, one HUT TripCode proves continuity: article memory, River history, ATLAS-7 boundaries, Gemini synthesis, monitor-next actions, and usage visibility.</p>
      <div class="goRow reveal" style="--delay:.54s">
        <a class="goButton" href="/demo/run" aria-label="Go to HUT proof"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M5 12h14"></path><path d="M13 6l6 6-6 6"></path></svg>Go To HUT Proof</a>
        <a class="goButton demoButton" href="/demo/run?autoplay=1" aria-label="Start automated two minute demo"><svg viewBox="0 0 24 24" aria-hidden="true"><path d="M8 5v14l11-7z"></path></svg>DEMO · 2 Min Sequence</a>
        <span class="pill">TF-SUB-9DA70A7F98</span>
      </div>
      <p class="sub reveal" style="--delay:.66s">This browser demo starts from the public DeltaSignal article, then runs the deployed Go Cloud Run surface used by A2A-compatible clients. The request path, response packet, and proof trail stay visible.</p>
      <span class="startCallout reveal" style="--delay:.76s">Apple is the flagship proof. HUT demonstrates reusable memory across time.</span>
      <div class="proofGrid reveal" style="--delay:.84s">
        <div class="card"><strong>1</strong><span>TripCode starts the loop</span></div>
        <div class="card"><strong>10</strong><span>HUT River nodes</span></div>
        <div class="card"><strong>A2A</strong><span>Agent artifact path</span></div>
      </div>
    </div>
    <div class="screenFrame reveal" style="--delay:.92s" aria-label="Animated research to logic pipeline">
      <div class="screenDots"><span></span><span></span><span></span></div>
      <div class="architectureStage">
        <svg class="archSvg" viewBox="0 0 980 520" role="img" aria-label="DeltaSignal agent architecture: User invokes Cloud Run A2A agent; MCP ATLAS-7, HUT River, and Evidence feed the agent; the A2A agent returns a research packet.">
          <defs>
            <marker id="arrowBlue" markerWidth="14" markerHeight="14" refX="12" refY="7" orient="auto">
              <path d="M0,0 L14,7 L0,14 Z" fill="#72d9ff"></path>
            </marker>
            <marker id="arrowGreen" markerWidth="16" markerHeight="16" refX="13" refY="8" orient="auto">
              <path d="M0,0 L16,8 L0,16 Z" fill="#5df58b"></path>
            </marker>
          </defs>
          <circle class="archPulse" cx="500" cy="172" r="210"></circle>
          <path class="archWire" marker-end="url(#arrowBlue)" d="M190 158 H350"></path>
          <path class="archWire" marker-end="url(#arrowBlue)" d="M742 158 H608"></path>
          <path class="archWire green" marker-end="url(#arrowGreen)" d="M500 222 V274"></path>
          <path class="archWire green" marker-end="url(#arrowGreen)" d="M310 362 C350 295 485 286 500 230"></path>
          <path class="archWire green" marker-end="url(#arrowGreen)" d="M690 362 C650 295 515 286 500 230"></path>
          <rect class="archNode primary" x="58" y="88" width="184" height="134" rx="25"></rect>
          <circle class="archIcon hot" cx="91" cy="140" r="5"></circle>
          <path class="archIcon hot" d="M82 154c3-8 15-8 18 0"></path>
          <text class="archLabel" x="128" y="147">User</text>
          <text class="archSub" x="128" y="184">intent + curl</text>
          <rect class="archNode primary" x="392" y="88" width="216" height="134" rx="25"></rect>
          <rect class="archIcon hot" x="417" y="137" width="15" height="11" rx="3"></rect>
          <path class="archIcon hot" d="M416 158h23"></path>
          <text class="archLabel" x="482" y="147">Cloud Run</text>
          <text class="archSub" x="482" y="184">A2A agent</text>
          <rect class="archNode primary" x="402" y="274" width="196" height="78" rx="22"></rect>
          <path class="archIcon hot" d="M424 313h18l5 5h26"></path>
          <path class="archIcon hot" d="M428 328h40"></path>
          <text class="archLabel" x="472" y="313">Output</text>
          <text class="archSub" x="452" y="338">Research packet</text>
          <rect class="archNode" x="748" y="88" width="174" height="134" rx="25"></rect>
          <circle class="archIcon blue" cx="775" cy="140" r="7"></circle>
          <path class="archIcon blue" d="M789 133h15M789 140h21M789 147h15"></path>
          <text class="archLabel" x="834" y="147">MCP</text>
          <text class="archSub" x="834" y="184">ATLAS-7</text>
          <rect class="archNode" x="192" y="362" width="260" height="98" rx="24"></rect>
          <path class="archIcon blue" d="M218 400h18l5 5h22"></path>
          <path class="archIcon blue" d="M218 416h34"></path>
          <text class="archLabel" x="298" y="405">HUT River</text>
          <text class="archSub" x="298" y="438">TripCode memory</text>
          <rect class="archNode" x="596" y="362" width="222" height="98" rx="24"></rect>
          <path class="archIcon blue" d="M616 400h16l5 5 5-5h19"></path>
          <path class="archIcon blue" d="M620 416h32"></path>
          <text class="archLabel" x="688" y="405">Evidence</text>
          <text class="archSub" x="688" y="438">SEC/XBRL refs</text>
        </svg>
      </div>
    </div>
  </section>
  <div class="landingWatermark">© 2026 AITrailblazer · DeltaSignal</div>
</main>
</body>
</html>`

func registerDemoUIRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /", serveRootLanding)
	mux.HandleFunc("GET /demo", serveDemoLanding)
	mux.HandleFunc("GET /demo/", serveDemoLanding)
	mux.HandleFunc("GET /demo/run", serveDemoUI)
	mux.HandleFunc("GET /demo/run/", serveDemoUI)
	mux.HandleFunc("GET /assets/deltasignal-app-icon.png", serveDeltaSignalLogo)
	mux.HandleFunc("GET /assets/substack-hut-article-top.png", serveSubstackHUTArticleTop)
	mux.HandleFunc("GET /assets/research-to-logic-pipeline.png", serveResearchToLogicPipeline)
}

func serveRootLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(rootLandingHTML))
}

func serveDemoLanding(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(demoLandingHTML))
}

func serveDemoUI(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	body := demoUIHTML
	if os.Getenv("DELTASIGNAL_DEMO_API_KEY") == "local-demo-key" {
		body = strings.Replace(body, `id="key" type="password" autocomplete="off" placeholder="DEMO_KEY"`, `id="key" type="password" autocomplete="off" placeholder="DEMO_KEY" value="local-demo-key"`, 1)
	}
	_, _ = w.Write([]byte(body))
}

func serveDeltaSignalLogo(w http.ResponseWriter, r *http.Request) {
	path := findDemoAsset("assets", "deltasignal-app-icon.png")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, path)
}

func serveSubstackHUTArticleTop(w http.ResponseWriter, r *http.Request) {
	path := findDemoAsset("DEMO", "assets", "substack-hut-article-top.png")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, path)
}

func serveResearchToLogicPipeline(w http.ResponseWriter, r *http.Request) {
	path := findDemoAsset("img", "RESEARCH-TO-LOGIC-PIPELINE.png")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	if _, err := os.Stat(path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	http.ServeFile(w, r, path)
}

func findDemoAsset(parts ...string) string {
	candidate := filepath.Join(parts...)
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	for _, prefix := range []string{"..", "../..", "../../.."} {
		candidate = filepath.Join(append([]string{prefix}, parts...)...)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}
