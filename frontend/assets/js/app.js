const sessionsList = document.getElementById("sessions-list");
const subjectsDatalist = document.getElementById("subjects");
const subjectsListEl = document.getElementById("subjects-list");
const totalMinutesEl = document.getElementById("total-minutes");
const sessionCountEl = document.getElementById("session-count");
const averageMinutesEl = document.getElementById("average-minutes");
const todayMinutesEl = document.getElementById("today-minutes");
const weekMinutesEl = document.getElementById("week-minutes");
const monthMinutesEl = document.getElementById("month-minutes");
const streakDaysEl = document.getElementById("streak-days");
const subjectBreakdownEl = document.getElementById("subject-breakdown");
const sessionForm = document.getElementById("session-form");
const sessionSubmitBtn = document.getElementById("session-submit");
const sessionCancelBtn = document.getElementById("session-cancel");
const sessionErrorEl = document.getElementById("session-error");
const subjectForm = document.getElementById("subject-form");
const subjectIdInput = document.getElementById("subject-id");
const subjectNameInput = document.getElementById("subject-name");
const subjectColorInput = document.getElementById("subject-color");
const subjectSubmitBtn = document.getElementById("subject-submit");
const subjectCancelBtn = document.getElementById("subject-cancel");
const subjectErrorEl = document.getElementById("subject-error");
const startTimeInput = document.getElementById("start-time");
const endTimeInput = document.getElementById("end-time");
const subjectChartCanvas = document.getElementById("subject-chart");
const trendChartCanvas = document.getElementById("trend-chart");
const navButtons = document.querySelectorAll(".app-nav__btn");
const views = document.querySelectorAll(".view");
const historySubjectFilter = document.getElementById("history-subject-filter");
const historyStartInput = document.getElementById("history-start-date");
const historyEndInput = document.getElementById("history-end-date");
const historyClearBtn = document.getElementById("history-clear-filters");
const historyTableBody = document.getElementById("history-table-body");
const loginForm = document.getElementById("login-form");
const registerForm = document.getElementById("register-form");
const loginErrorEl = document.getElementById("login-error");
const registerErrorEl = document.getElementById("register-error");
const googleLoginBtn = document.getElementById("google-login-btn");
const logoutBtn = document.getElementById("logout-btn");
const showRegisterBtn = document.getElementById("show-register-btn");
const showLoginBtn = document.getElementById("show-login-btn");
const liveTrackSubjectInput = document.getElementById("live-track-subject");
const liveTrackColorInput = document.getElementById("live-track-color");
const liveTrackTimerEl = document.getElementById("live-track-timer");
const liveTrackStatusEl = document.getElementById("live-track-status");
const liveTrackStartBtn = document.getElementById("live-track-start");
const liveTrackPauseBtn = document.getElementById("live-track-pause");
const liveTrackTriggerBtn = document.getElementById("live-track-trigger");
const liveTrackLogBtn = document.getElementById("live-track-log");
const liveTrackCircle = document.getElementById("live-track-circle");
const liveTrackTimerWrapper = document.getElementById("live-track-timer-wrapper");
const liveTrackModeEl = document.getElementById("live-track-mode");
const liveTrackModeButtons = Array.from(document.querySelectorAll("[data-live-track-mode]"));
const liveTrackSelectedModeEl = document.getElementById("live-track-selected-mode");
const liveTrackChangeModeBtn = document.getElementById("live-track-change-mode");
const liveTrackDurationWrapper = document.getElementById("live-track-duration-wrapper");
const liveTrackDurationInput = document.getElementById("live-track-duration");
const liveTrackSetupEl = document.getElementById("live-track-setup");
const liveTrackControlsEl = document.getElementById("live-track-controls");
const liveTrackMessageEl = document.getElementById("live-track-message");
const liveTrackCardEl = document.querySelector(".live-track-card");
const viewHistoryBtn = document.getElementById("view-history-btn");

if (typeof Chart !== "undefined") {
  const rootStyles = getComputedStyle(document.documentElement);
  Chart.defaults.color = rootStyles.color || "#e2e8f0";
  Chart.defaults.font.family = rootStyles.getPropertyValue("--font-family") || getComputedStyle(document.body).fontFamily;
}

let sessions = [];
let subjects = [];
let summaryData = null;
let editingSessionId = null;
let editingSubjectId = null;
let subjectChart = null;
let trendChart = null;
let currentUser = null;
let isAuthenticated = false;
let dataLoaded = false;
let pendingView = localStorage.getItem("activeView") || "dashboard";

const defaultSubjectColor = liveTrackColorInput?.value || "#6366f1";
const LIVE_TRACK_MIN_MS = 60 * 1000;

const liveTrackState = {
  status: "idle",
  subject: "",
  color: defaultSubjectColor,
  startedAt: null,
  elapsedMs: 0,
  timerId: null,
  isSubmitting: false,
  showSetup: false,
  showModePicker: false,
  mode: null,
  timerDurationMs: 0,
};

const fallbackColors = [
  "#6366f1",
  "#38bdf8",
  "#f472b6",
  "#facc15",
  "#34d399",
  "#f97316",
  "#22d3ee",
  "#a855f7",
];

function setDefaultTimes() {
  const now = new Date();
  const endISO = toLocalInputValue(now);
  const startISO = toLocalInputValue(new Date(now.getTime() - 60 * 60 * 1000));
  startTimeInput.value = startISO;
  endTimeInput.value = endISO;
}

function toLocalInputValue(date) {
  const pad = (num) => String(num).padStart(2, "0");
  const year = date.getFullYear();
  const month = pad(date.getMonth() + 1);
  const day = pad(date.getDate());
  const hours = pad(date.getHours());
  const minutes = pad(date.getMinutes());
  return `${year}-${month}-${day}T${hours}:${minutes}`;
}

function escapeHTML(value) {
  return value
    ? value
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;")
    : "";
}

async function fetchJSON(url, options = {}) {
  const mergedOptions = {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      ...(options && options.headers ? options.headers : {}),
    },
    ...options,
  };

  // Avoid sending Content-Type for GET requests without body
  if (!mergedOptions.body && (!options || !options.headers || !options.headers["Content-Type"])) {
    delete mergedOptions.headers["Content-Type"];
  }

  const response = await fetch(url, mergedOptions);
  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || "Request failed");
  }
  if (response.status === 204) {
    return null;
  }
  return response.json();
}

function setLiveTrackAccent(color) {
  if (!liveTrackCardEl) return;
  const normalized = color && /^#([0-9a-f]{3}|[0-9a-f]{6})$/i.test(color) ? color : defaultSubjectColor;
  liveTrackCardEl.style.setProperty("--live-track-accent", normalized);
}

function formatDuration(ms) {
  if (!Number.isFinite(ms) || ms <= 0) {
    return "00:00:00";
  }
  const totalSeconds = Math.floor(ms / 1000);
  const hours = String(Math.floor(totalSeconds / 3600)).padStart(2, "0");
  const minutes = String(Math.floor((totalSeconds % 3600) / 60)).padStart(2, "0");
  const seconds = String(totalSeconds % 60).padStart(2, "0");
  return `${hours}:${minutes}:${seconds}`;
}

function getLiveTrackElapsedRawMs() {
  let elapsed = liveTrackState.elapsedMs;
  if (liveTrackState.status === "running" && liveTrackState.startedAt) {
    elapsed += Date.now() - liveTrackState.startedAt;
  }
  return elapsed;
}

function getLiveTrackElapsedMs() {
  const raw = getLiveTrackElapsedRawMs();
  if (liveTrackState.mode === "timer" && liveTrackState.timerDurationMs > 0) {
    return Math.min(raw, liveTrackState.timerDurationMs);
  }
  return raw;
}

function getLiveTrackRemainingMs() {
  if (liveTrackState.mode !== "timer" || liveTrackState.timerDurationMs <= 0) {
    return null;
  }
  return Math.max(liveTrackState.timerDurationMs - getLiveTrackElapsedRawMs(), 0);
}

function clearLiveTrackInterval() {
  if (liveTrackState.timerId) {
    window.clearInterval(liveTrackState.timerId);
    liveTrackState.timerId = null;
  }
}

function startLiveTrackInterval() {
  clearLiveTrackInterval();
  liveTrackState.timerId = window.setInterval(handleLiveTrackTick, 1000);
  handleLiveTrackTick();
}

function updateLiveTrackTimerText() {
  if (!liveTrackTimerEl) return;
  const displayMs =
    liveTrackState.mode === "timer" && liveTrackState.timerDurationMs > 0
      ? getLiveTrackRemainingMs() ?? 0
      : getLiveTrackElapsedMs();
  liveTrackTimerEl.textContent = formatDuration(displayMs);
}

function handleLiveTrackTick() {
  updateLiveTrackTimerText();
  if (
    liveTrackState.mode === "timer" &&
    liveTrackState.status === "running" &&
    (getLiveTrackRemainingMs() ?? 0) <= 0
  ) {
    captureLiveTrackElapsed();
    liveTrackState.status = "paused";
    clearLiveTrackInterval();
    showMessage(liveTrackMessageEl, "Timer finished! Tap Log to save.");
    updateLiveTrackUI();
  }
}

function updateLiveTrackUI() {
  const { status, isSubmitting, showSetup, showModePicker, mode } = liveTrackState;
  const hasProgress = getLiveTrackElapsedMs() > 0;
  const showTimer = status !== "idle";
  const isTimerMode = mode === "timer";

  if (liveTrackTriggerBtn) {
    liveTrackTriggerBtn.disabled = status !== "idle" || isSubmitting;
    liveTrackTriggerBtn.classList.toggle("hidden", showTimer);
    const expanded = showModePicker || showSetup;
    liveTrackTriggerBtn.setAttribute("aria-pressed", String(expanded));
    liveTrackTriggerBtn.setAttribute("aria-expanded", String(expanded));
  }

  if (liveTrackCircle) {
    liveTrackCircle.classList.toggle("is-active", showTimer);
    liveTrackCircle.classList.toggle("is-paused", status === "paused");
  }

  liveTrackTimerWrapper?.classList.toggle("hidden", !showTimer);
  liveTrackModeEl?.classList.toggle("hidden", !showModePicker);
  liveTrackSetupEl?.classList.toggle("hidden", !showSetup);
  liveTrackControlsEl?.classList.toggle("hidden", !showTimer);
  liveTrackDurationWrapper?.classList.toggle("hidden", !showSetup || !isTimerMode);

  if (liveTrackStartBtn) {
    const canStart = showSetup && !isSubmitting && status === "idle" && Boolean(mode);
    liveTrackStartBtn.disabled = !canStart;
  }
  if (liveTrackPauseBtn) {
    liveTrackPauseBtn.disabled = status === "idle" || isSubmitting;
    liveTrackPauseBtn.textContent = status === "paused" ? "Resume" : "Pause";
  }
  if (liveTrackLogBtn) {
    liveTrackLogBtn.disabled = !hasProgress || isSubmitting;
  }

  if (liveTrackSelectedModeEl) {
    liveTrackSelectedModeEl.textContent = mode === "timer" ? "Timer" : "Stopwatch";
    liveTrackSelectedModeEl.classList.toggle("hidden", !mode);
  }
  liveTrackChangeModeBtn?.classList.toggle("hidden", !mode || status !== "idle");

  if (liveTrackModeButtons.length) {
    liveTrackModeButtons.forEach((button) => {
      const buttonMode = button.dataset.liveTrackMode;
      button.classList.toggle("active", buttonMode === mode);
    });
  }

  if (liveTrackStatusEl) {
    if (isSubmitting) {
      liveTrackStatusEl.textContent = "Saving session…";
    } else if (status === "running") {
      liveTrackStatusEl.textContent =
        mode === "timer" ? "Timer in progress." : "Stopwatch running.";
    } else if (status === "paused" && hasProgress) {
      liveTrackStatusEl.textContent = "Paused — resume or log below.";
    } else if (showModePicker) {
      liveTrackStatusEl.textContent = "Choose stopwatch or timer to continue.";
    } else if (showSetup && mode === "timer") {
      liveTrackStatusEl.textContent = "Set your timer details, then press Start.";
    } else if (showSetup) {
      liveTrackStatusEl.textContent = "Name your stopwatch session and press Start.";
    } else {
      liveTrackStatusEl.textContent = "Tap Track Live to begin.";
    }
  }

  updateLiveTrackTimerText();
}

function handleTrackTriggerClick() {
  if (liveTrackState.status !== "idle" || liveTrackState.isSubmitting) {
    return;
  }
  if (!liveTrackState.showModePicker && !liveTrackState.showSetup && !liveTrackState.mode) {
    liveTrackState.showModePicker = true;
  } else {
    resetLiveTrackState();
  }
  updateLiveTrackUI();
}

function selectLiveTrackMode(mode) {
  if (liveTrackState.status !== "idle" || liveTrackState.isSubmitting) {
    return;
  }
  liveTrackState.mode = mode;
  liveTrackState.showModePicker = false;
  liveTrackState.showSetup = true;
  if (mode === "timer") {
    liveTrackState.timerDurationMs = getTimerDurationMsFromInput();
  } else {
    liveTrackState.timerDurationMs = 0;
  }
  showMessage(liveTrackMessageEl, "");
  updateLiveTrackUI();
  window.requestAnimationFrame(() => liveTrackSubjectInput?.focus());
}

function handleLiveTrackChangeMode() {
  if (liveTrackState.status !== "idle" || liveTrackState.isSubmitting) {
    return;
  }
  liveTrackState.showSetup = false;
  liveTrackState.showModePicker = true;
  updateLiveTrackUI();
}

function getTimerDurationMsFromInput() {
  if (!liveTrackDurationInput) {
    return 0;
  }
  const raw = (liveTrackDurationInput.value || "").trim();
  if (!raw) {
    return 0;
  }
  const digitsOnly = raw.replace(/[^0-9]/g, "");
  if (digitsOnly !== raw) {
    liveTrackDurationInput.value = digitsOnly;
  }
  if (!digitsOnly) {
    return 0;
  }
  const value = Number(digitsOnly);
  if (!Number.isFinite(value)) {
    return 0;
  }
  const clamped = Math.min(Math.max(value, 1), 240);
  if (clamped !== value) {
    liveTrackDurationInput.value = String(clamped);
  }
  return clamped * 60 * 1000;
}

function handleTimerDurationInput() {
  if (!liveTrackDurationInput) {
    return;
  }
  liveTrackDurationInput.value = liveTrackDurationInput.value.replace(/[^0-9]/g, "");
  if (liveTrackState.mode !== "timer") {
    return;
  }
  if (liveTrackState.status === "idle") {
    liveTrackState.timerDurationMs = getTimerDurationMsFromInput();
  }
}

function resetLiveTrackState(options = {}) {
  const { preserveMessage = false } = options;
  clearLiveTrackInterval();
  liveTrackState.status = "idle";
  liveTrackState.subject = "";
  liveTrackState.color = defaultSubjectColor;
  liveTrackState.startedAt = null;
  liveTrackState.elapsedMs = 0;
  liveTrackState.showSetup = false;
  liveTrackState.showModePicker = false;
  liveTrackState.mode = null;
  liveTrackState.timerDurationMs = 0;
  if (liveTrackSubjectInput) {
    liveTrackSubjectInput.value = "";
  }
  if (liveTrackColorInput) {
    liveTrackColorInput.value = defaultSubjectColor;
  }
  if (liveTrackDurationInput && liveTrackDurationInput.defaultValue) {
    liveTrackDurationInput.value = liveTrackDurationInput.defaultValue;
  }
  if (!preserveMessage) {
    showMessage(liveTrackMessageEl, "");
  }
  setLiveTrackAccent(defaultSubjectColor);
  updateLiveTrackUI();
}

function captureLiveTrackElapsed() {
  if (liveTrackState.startedAt) {
    liveTrackState.elapsedMs += Date.now() - liveTrackState.startedAt;
    liveTrackState.startedAt = null;
  }
}

function handleLiveTrackStart() {
  if (liveTrackState.status !== "idle" || liveTrackState.isSubmitting) {
    return;
  }
  if (!liveTrackState.mode) {
    showMessage(liveTrackMessageEl, "Pick stopwatch or timer first.");
    return;
  }
  const subject = liveTrackSubjectInput?.value?.trim() || "";
  if (!subject) {
    showMessage(liveTrackMessageEl, "Subject is required to start tracking.");
    liveTrackSubjectInput?.focus();
    return;
  }
  if (liveTrackState.mode === "timer") {
    const durationMs = getTimerDurationMsFromInput();
    if (durationMs <= 0) {
      showMessage(liveTrackMessageEl, "Timer duration must be at least 1 minute.");
      liveTrackDurationInput?.focus();
      return;
    }
    liveTrackState.timerDurationMs = durationMs;
  } else {
    liveTrackState.timerDurationMs = 0;
  }
  liveTrackState.subject = subject;
  liveTrackState.color = liveTrackColorInput?.value || defaultSubjectColor;
  liveTrackState.elapsedMs = 0;
  liveTrackState.startedAt = Date.now();
  liveTrackState.status = "running";
  liveTrackState.showSetup = false;
  showMessage(liveTrackMessageEl, "");
  setLiveTrackAccent(liveTrackState.color);
  startLiveTrackInterval();
  updateLiveTrackUI();
}

function toggleLiveTrackPause() {
  if (liveTrackState.isSubmitting) {
    return;
  }
  if (liveTrackState.status === "running") {
    captureLiveTrackElapsed();
    liveTrackState.status = "paused";
    clearLiveTrackInterval();
  } else if (liveTrackState.status === "paused") {
    liveTrackState.startedAt = Date.now();
    liveTrackState.status = "running";
    startLiveTrackInterval();
  } else {
    return;
  }
  showMessage(liveTrackMessageEl, "");
  updateLiveTrackUI();
}

async function handleLiveTrackLog() {
  if (liveTrackState.status === "idle" || liveTrackState.isSubmitting) {
    return;
  }
  const wasRunning = liveTrackState.status === "running";
  captureLiveTrackElapsed();
  clearLiveTrackInterval();
  if (wasRunning) {
    liveTrackState.status = "paused";
  }
  updateLiveTrackUI();

  const totalMs = getLiveTrackElapsedMs();
  if (totalMs < LIVE_TRACK_MIN_MS) {
    showMessage(liveTrackMessageEl, "Track at least one minute before saving.");
    if (wasRunning) {
      liveTrackState.startedAt = Date.now();
      liveTrackState.status = "running";
      startLiveTrackInterval();
    }
    updateLiveTrackUI();
    return;
  }

  liveTrackState.isSubmitting = true;
  updateLiveTrackUI();
  showMessage(liveTrackMessageEl, "");

  try {
    await ensureSubjectExists(liveTrackState.subject, liveTrackState.color);
    const endTime = new Date();
    const startTime = new Date(endTime.getTime() - totalMs);
    const payload = {
      subject: liveTrackState.subject,
      notes: "Logged via Live Track",
      reflection: "",
      startTime: startTime.toISOString(),
      endTime: endTime.toISOString(),
    };
    await fetchJSON("/api/study-sessions", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    showMessage(liveTrackMessageEl, "Session saved to your log.");
    resetLiveTrackState({ preserveMessage: true });
    await Promise.all([loadSessions(), loadSummary()]);
  } catch (error) {
    console.error("Failed to save live track session", error);
    showMessage(liveTrackMessageEl, error.message || "Failed to save session.");
    liveTrackState.status = "paused";
  } finally {
    liveTrackState.isSubmitting = false;
    updateLiveTrackUI();
  }
}

function initLiveTrack() {
  if (!liveTrackStartBtn || !liveTrackTriggerBtn) return;
  liveTrackTriggerBtn.addEventListener("click", handleTrackTriggerClick);
  liveTrackStartBtn.addEventListener("click", handleLiveTrackStart);
  liveTrackPauseBtn?.addEventListener("click", toggleLiveTrackPause);
  liveTrackLogBtn?.addEventListener("click", handleLiveTrackLog);
  if (liveTrackModeButtons.length) {
    liveTrackModeButtons.forEach((button) => {
      button.addEventListener("click", () => selectLiveTrackMode(button.dataset.liveTrackMode));
    });
  }
  liveTrackChangeModeBtn?.addEventListener("click", handleLiveTrackChangeMode);
  liveTrackColorInput?.addEventListener("input", handleLiveTrackColorPreview);
  liveTrackColorInput?.addEventListener("change", handleLiveTrackColorPreview);
  liveTrackDurationInput?.addEventListener("input", handleTimerDurationInput);
  liveTrackDurationInput?.addEventListener("change", handleTimerDurationInput);
  liveTrackDurationInput?.addEventListener("blur", handleTimerDurationInput);
  resetLiveTrackState();
}

async function loadSubjects() {
  try {
    subjects = await fetchJSON("/api/subjects");
    renderSubjects();
    syncSubjectOptions();
    if (summaryData) {
      renderSummary();
    }
  } catch (error) {
    console.error("Failed to load subjects", error);
    showMessage(subjectErrorEl, "Failed to load subjects.");
  }
}

async function loadSessions() {
  try {
    sessions = await fetchJSON("/api/study-sessions");
    renderSessions();
    renderHistory();
    syncSubjectOptions();
  } catch (error) {
    console.error("Failed to load sessions", error);
    showMessage(sessionErrorEl, "Failed to load study sessions.");
  }
}

async function loadSummary() {
  try {
    summaryData = await fetchJSON("/api/progress/summary");
    renderSummary();
  } catch (error) {
    console.error("Failed to load summary", error);
  }
}

function renderSubjects() {
  if (!subjects.length) {
    subjectsListEl.innerHTML = "<li>No subjects yet. Add one above.</li>";
    return;
  }

  subjectsListEl.innerHTML = subjects
    .map(
      (subject) => `
        <li data-id="${subject.id}">
          <div class="subject-info">
            <span class="subject-color" style="background-color: ${
              subject.color || "#6366f1"
            }"></span>
            <span>${escapeHTML(subject.name)}</span>
          </div>
          <div class="subject-actions">
            <button type="button" class="secondary edit-subject" data-id="${subject.id}">Edit</button>
            <button type="button" class="secondary delete-subject" data-id="${subject.id}">Delete</button>
          </div>
        </li>
      `
    )
    .join("");
}

function renderSessions() {
  if (!sessionsList) return;
  if (!Array.isArray(sessions) || sessions.length === 0) {
    sessionsList.innerHTML = "<li>No study sessions yet.</li>";
    return;
  }

  const recentSessions = sessions.slice(0, 5);
  let markup = recentSessions
    .map((session) => {
      const start = formatDate(session.startTime);
      const end = formatDate(session.endTime);
      return `
        <li data-id="${session.id}">
          <strong>${escapeHTML(session.subject)}</strong>
          <div class="meta">
            ${start} - ${end} · ${session.durationMinutes} minutes
          </div>
          <div class="meta">${escapeHTML(session.notes) || "No notes"}</div>
          ${
            session.reflection
              ? `<div class="meta italic">Reflection: ${escapeHTML(session.reflection)}</div>`
              : ""
          }
          <div class="session-actions">
            <button type="button" class="secondary edit-session" data-id="${session.id}">Edit</button>
            <button type="button" class="secondary delete-session" data-id="${session.id}">Delete</button>
          </div>
        </li>
      `;
    })
    .join("");

  if (sessions.length > recentSessions.length) {
    markup += `<li class="sessions-footnote">View earlier entries in the History tab.</li>`;
  }

  sessionsList.innerHTML = markup;
}

function renderSummary(summary = summaryData) {
  if (!summary) {
    totalMinutesEl.textContent = "0";
    sessionCountEl.textContent = "0";
    averageMinutesEl.textContent = "0.0";
    todayMinutesEl.textContent = "0";
    weekMinutesEl.textContent = "0";
    monthMinutesEl.textContent = "0";
    streakDaysEl.textContent = "0";
    subjectBreakdownEl.innerHTML = "<p>No data yet.</p>";
    updateSubjectChart({});
    updateTrendChart([]);
    return;
  }

  totalMinutesEl.textContent = summary.totalMinutes ?? 0;
  sessionCountEl.textContent = summary.sessionCount ?? 0;
  const avg =
    summary.sessionCount > 0 && Number.isFinite(summary.averageSessionMinutes)
      ? summary.averageSessionMinutes
      : 0;
  averageMinutesEl.textContent = avg.toFixed(1);
  todayMinutesEl.textContent = summary.todayMinutes ?? 0;
  weekMinutesEl.textContent = summary.weekMinutes ?? 0;
  monthMinutesEl.textContent = summary.monthMinutes ?? 0;
  streakDaysEl.textContent = summary.streakDays ?? 0;

  renderSubjectBreakdown(summary.bySubject || {});
  updateSubjectChart(summary.bySubject || {});
  updateTrendChart(summary.dailyTrend || []);
}

function renderSubjectBreakdown(bySubject) {
  const entries = Object.entries(bySubject || {});
  if (!entries.length) {
    subjectBreakdownEl.innerHTML = "<p>No data yet.</p>";
    return;
  }

  const totalMinutes = entries.reduce((acc, [, minutes]) => acc + minutes, 0);

  subjectBreakdownEl.innerHTML = entries
    .sort((a, b) => b[1] - a[1])
    .map(([name, minutes], index) => {
      const color = getSubjectColor(name, index);
      const percentage = totalMinutes ? Math.round((minutes / totalMinutes) * 100) : 0;
      return `
        <div class="subject-breakdown-item">
          <div class="subject-breakdown-info">
            <span class="subject-dot" style="background-color: ${color}"></span>
            <span>${escapeHTML(name)}</span>
          </div>
          <span>${minutes} min (${percentage}%)</span>
        </div>
      `;
    })
    .join("");
}

function updateSubjectChart(bySubject) {
  if (!subjectChartCanvas || typeof Chart === "undefined") {
    return;
  }

  const entries = Object.entries(bySubject || {});
  if (subjectChart) {
    subjectChart.destroy();
    subjectChart = null;
  }
  if (!entries.length) {
    return;
  }

  const labels = entries.map(([name]) => name);
  const data = entries.map(([, minutes]) => minutes);
  const colors = labels.map((label, index) => getSubjectColor(label, index));

  subjectChart = new Chart(subjectChartCanvas, {
    type: "doughnut",
    data: {
      labels,
      datasets: [
        {
          data,
          backgroundColor: colors,
          borderWidth: 0,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      plugins: {
        legend: { display: false },
        tooltip: {
          callbacks: {
            label(context) {
              const label = context.label || "";
              const value = context.parsed || 0;
              return `${label}: ${value} min`;
            },
          },
        },
      },
      interaction: { intersect: false },
      cutout: "62%",
    },
  });
}

function updateTrendChart(trend) {
  if (!trendChartCanvas || typeof Chart === "undefined") {
    return;
  }

  if (trendChart) {
    trendChart.destroy();
    trendChart = null;
  }

  if (!Array.isArray(trend) || trend.length === 0) {
    return;
  }

  const labels = trend.map((entry) => formatTrendLabel(entry.date));
  const data = trend.map((entry) => entry.totalMinutes ?? 0);

  trendChart = new Chart(trendChartCanvas, {
    type: "line",
    data: {
      labels,
      datasets: [
        {
          label: "Minutes",
          data,
          borderColor: "#6366f1",
          backgroundColor: "#6366f1",
          fill: false,
          tension: 0.3,
          pointRadius: 3,
          pointHoverRadius: 5,
        },
      ],
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      animation: false,
      interaction: {
        mode: "index",
        intersect: false,
      },
      plugins: {
        legend: { display: false },
        tooltip: {
          displayColors: false,
        },
      },
      scales: {
        y: {
          beginAtZero: true,
          grid: { color: "rgba(148, 163, 184, 0.12)" },
          ticks: { stepSize: 10 },
        },
        x: {
          grid: { display: false },
        },
      },
    },
  });
}

function syncSubjectOptions() {
  const nameSet = new Set(subjects.map((subject) => subject.name).filter(Boolean));
  sessions.forEach((session) => {
    if (session.subject) {
      nameSet.add(session.subject);
    }
  });

  subjectsDatalist.innerHTML = Array.from(nameSet)
    .sort((a, b) => a.localeCompare(b))
    .map((name) => `<option value="${escapeHTML(name)}"></option>`)
    .join("");

  updateHistorySubjectOptions(Array.from(nameSet));
}

function getSubjectColor(name, index) {
  const match = subjects.find(
    (subject) => subject.name && subject.name.toLowerCase() === name.toLowerCase()
  );
  if (match && match.color) {
    return match.color;
  }
  return fallbackColors[index % fallbackColors.length];
}

function formatDate(value) {
  if (!value) return "Unknown";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Unknown";
  return date.toLocaleString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    month: "short",
    day: "numeric",
  });
}

function formatHistoryDate(value) {
  if (!value) return "Unknown";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "Unknown";
  return date.toLocaleString(undefined, {
    weekday: "short",
    month: "short",
    day: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

function formatTrendLabel(dateString) {
  if (!dateString) return "";
  const [year, month, day] = dateString.split("-").map(Number);
  const date = new Date(year, (month || 1) - 1, day || 1);
  return date.toLocaleDateString(undefined, { month: "short", day: "numeric" });
}

function showMessage(element, message = "") {
  if (!element) return;
  if (!message) {
    element.textContent = "";
    element.classList.add("hidden");
  } else {
    element.textContent = message;
    element.classList.remove("hidden");
  }
}

function setAuthenticated(user) {
  currentUser = user;
  isAuthenticated = Boolean(user);
  document.body.classList.toggle("authed", isAuthenticated);

  if (!isAuthenticated) {
    dataLoaded = false;
    resetLiveTrackState();
    setAuthMode("login");
    activateView("auth", { skipSave: true, force: true });
  } else {
    const targetView = pendingView || "dashboard";
    activateView(targetView, { force: true });
  }
}

function setAuthMode(mode) {
  if (!loginForm || !registerForm) return;
  const showRegister = mode === "register";
  loginForm.classList.toggle("hidden", showRegister);
  registerForm.classList.toggle("hidden", !showRegister);
}

async function loadCurrentUser() {
  try {
    const user = await fetchJSON("/api/auth/me", { headers: {} });
    setAuthenticated(user);
  } catch (error) {
    setAuthenticated(null);
  }
}

async function loadAuthedData() {
  if (!isAuthenticated) return;
  await loadSubjects();
  await Promise.all([loadSessions(), loadSummary()]);
  dataLoaded = true;
}

function updateHistorySubjectOptions(subjectNames) {
  if (!historySubjectFilter) return;
  const existing = historySubjectFilter.value;
  historySubjectFilter.innerHTML =
    '<option value="all">All subjects</option>' +
    subjectNames
      .sort((a, b) => a.localeCompare(b))
      .map((name) => `<option value="${escapeHTML(name)}">${escapeHTML(name)}</option>`)
      .join("");
  if (existing && subjectNames.includes(existing)) {
    historySubjectFilter.value = existing;
  }
}

function getFilteredSessions() {
  let filtered = [...sessions];
  const subjectFilter = historySubjectFilter?.value || "all";
  const startDate = historyStartInput?.value ? new Date(historyStartInput.value) : null;
  const endDate = historyEndInput?.value ? new Date(historyEndInput.value) : null;

  if (subjectFilter !== "all") {
    filtered = filtered.filter(
      (session) => session.subject && session.subject.toLowerCase() === subjectFilter.toLowerCase()
    );
  }

  if (startDate) {
    filtered = filtered.filter((session) => {
      const start = new Date(session.startTime);
      return start >= startDate;
    });
  }

  if (endDate) {
    const endOfDay = new Date(endDate);
    endOfDay.setHours(23, 59, 59, 999);
    filtered = filtered.filter((session) => {
      const start = new Date(session.startTime);
      return start <= endOfDay;
    });
  }

  return filtered;
}

function renderHistory() {
  if (!historyTableBody) return;
  const filtered = getFilteredSessions();

  if (!filtered.length) {
    historyTableBody.innerHTML =
      '<tr><td colspan="4" class="history-empty">No sessions match these filters.</td></tr>';
    return;
  }

  historyTableBody.innerHTML = filtered
    .map((session) => {
      const cleanNotes = escapeHTML(session.notes || "—");
      return `
        <tr>
          <td>${escapeHTML(session.subject || "Unknown")}</td>
          <td>${session.durationMinutes || 0} min</td>
          <td>${formatHistoryDate(session.startTime)}</td>
          <td>${cleanNotes}</td>
        </tr>
      `;
    })
    .join("");
}

async function handleSessionSubmit(event) {
  event.preventDefault();
  showMessage(sessionErrorEl, "");

  const subject = document.getElementById("subject").value.trim();
  const notes = document.getElementById("notes").value.trim();
  const reflection = document.getElementById("reflection").value.trim();
  const startTime = startTimeInput.value;
  const endTime = endTimeInput.value;

  if (!subject || !startTime || !endTime) {
    showMessage(sessionErrorEl, "Please fill subject, start, and end time.");
    return;
  }

  const payload = {
    subject,
    notes,
    reflection,
    startTime: new Date(startTime).toISOString(),
    endTime: new Date(endTime).toISOString(),
  };

  const url = editingSessionId
    ? `/api/study-sessions/${editingSessionId}`
    : "/api/study-sessions";

  const method = editingSessionId ? "PUT" : "POST";

  try {
    await fetchJSON(url, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    resetSessionForm();
    await Promise.all([loadSessions(), loadSummary()]);
  } catch (error) {
    console.error("Failed to submit session", error);
    showMessage(sessionErrorEl, error.message);
  }
}

async function handleSubjectSubmit(event) {
  event.preventDefault();
  showMessage(subjectErrorEl, "");

  const name = subjectNameInput.value.trim();
  const color = subjectColorInput.value.trim();

  if (!name) {
    showMessage(subjectErrorEl, "Subject name is required.");
    return;
  }

  const payload = { name, color };
  let url = "/api/subjects";
  let method = "POST";

  if (editingSubjectId) {
    url = `/api/subjects/${editingSubjectId}`;
    method = "PUT";
    payload.id = editingSubjectId;
  }

  try {
    await fetchJSON(url, {
      method,
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });

    resetSubjectForm();
    await loadSubjects();
  } catch (error) {
    console.error("Failed to submit subject", error);
    showMessage(subjectErrorEl, error.message);
  }
}

async function ensureSubjectExists(name, color) {
  const normalized = name?.trim().toLowerCase();
  if (!normalized) {
    throw new Error("Subject name missing.");
  }

  const existing = subjects.find(
    (subject) => subject.name && subject.name.toLowerCase() === normalized
  );
  if (existing) {
    return existing;
  }

  try {
    await fetchJSON("/api/subjects", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, color }),
    });
    await loadSubjects();
  } catch (error) {
    const message = (error.message || "").toLowerCase();
    if (message.includes("exists")) {
      await loadSubjects();
    } else {
      throw error;
    }
  }

  return subjects.find(
    (subject) => subject.name && subject.name.toLowerCase() === normalized
  );
}

function resetSessionForm() {
  sessionForm.reset();
  editingSessionId = null;
  sessionSubmitBtn.textContent = "Save Session";
  sessionCancelBtn.classList.add("hidden");
  showMessage(sessionErrorEl, "");
  setDefaultTimes();
}

function resetSubjectForm() {
  subjectForm.reset();
  subjectColorInput.value = "#6366f1";
  subjectIdInput.value = "";
  editingSubjectId = null;
  subjectSubmitBtn.textContent = "Add Subject";
  subjectCancelBtn.classList.add("hidden");
  showMessage(subjectErrorEl, "");
}

function activateView(name, options = {}) {
  const { skipSave = false, force = false } = options;
  if (!views.length) {
    return;
  }

  let targetView = name;

  if (!force) {
    if (!isAuthenticated && name !== "auth") {
      pendingView = name;
      targetView = "auth";
    }
    if (isAuthenticated && name === "auth") {
      targetView = pendingView || "dashboard";
    }
  }

  if (!Array.from(views).some((view) => view.dataset.view === targetView)) {
    targetView = views[0].dataset.view;
  }

  views.forEach((view) => {
    view.classList.toggle("active", view.dataset.view === targetView);
  });

  navButtons.forEach((btn) => {
    const isActive = btn.dataset.view === targetView;
    btn.classList.toggle("active", isActive);
    btn.setAttribute("aria-selected", String(isActive));
  });

  if (!skipSave && targetView !== "auth") {
    pendingView = targetView;
    localStorage.setItem("activeView", targetView);
  }
}

function beginSessionEdit(id) {
  const session = sessions.find((item) => item.id === id);
  if (!session) return;

  editingSessionId = id;
  document.getElementById("subject").value = session.subject || "";
  document.getElementById("notes").value = session.notes || "";
  document.getElementById("reflection").value = session.reflection || "";

  if (session.startTime) {
    startTimeInput.value = toLocalInputValue(new Date(session.startTime));
  }
  if (session.endTime) {
    endTimeInput.value = toLocalInputValue(new Date(session.endTime));
  }

  sessionSubmitBtn.textContent = "Update Session";
  sessionCancelBtn.classList.remove("hidden");
}

function beginSubjectEdit(id) {
  const subject = subjects.find((item) => item.id === id);
  if (!subject) return;

  editingSubjectId = id;
  subjectIdInput.value = id;
  subjectNameInput.value = subject.name || "";
  subjectColorInput.value = subject.color || "#6366f1";

  subjectSubmitBtn.textContent = "Update Subject";
  subjectCancelBtn.classList.remove("hidden");
}

async function deleteSession(id) {
  if (!window.confirm("Delete this study session?")) return;
  try {
    await fetchJSON(`/api/study-sessions/${id}`, { method: "DELETE" });
    await Promise.all([loadSessions(), loadSummary()]);
  } catch (error) {
    console.error("Failed to delete session", error);
    showMessage(sessionErrorEl, "Failed to delete session.");
  }
}

async function deleteSubject(id) {
  if (
    !window.confirm(
      "Delete this subject? Existing sessions will remain unchanged."
    )
  ) {
    return;
  }
  try {
    await fetchJSON(`/api/subjects/${id}`, { method: "DELETE" });
    if (editingSubjectId === id) {
      resetSubjectForm();
    }
    await loadSubjects();
  } catch (error) {
    console.error("Failed to delete subject", error);
    showMessage(subjectErrorEl, "Failed to delete subject.");
  }
}

sessionForm.addEventListener("submit", handleSessionSubmit);
sessionCancelBtn.addEventListener("click", resetSessionForm);

subjectForm.addEventListener("submit", handleSubjectSubmit);
subjectCancelBtn.addEventListener("click", resetSubjectForm);

sessionsList.addEventListener("click", (event) => {
  const target = event.target;
  if (!(target instanceof HTMLElement)) return;
  const id = target.dataset.id;
  if (!id) return;

  if (target.classList.contains("edit-session")) {
    beginSessionEdit(id);
  } else if (target.classList.contains("delete-session")) {
    deleteSession(id);
  }
});

subjectsListEl.addEventListener("click", (event) => {
  const target = event.target;
  if (!(target instanceof HTMLElement)) return;
  const id = target.dataset.id;
  if (!id) return;

  if (target.classList.contains("edit-subject")) {
    beginSubjectEdit(id);
  } else if (target.classList.contains("delete-subject")) {
    deleteSubject(id);
  }
});

const savedView = localStorage.getItem("activeView") || "dashboard";
activateView(savedView);

navButtons.forEach((btn) => {
  btn.addEventListener("click", () => {
    if (!isAuthenticated) return;
    activateView(btn.dataset.view);
  });
});

if (viewHistoryBtn) {
  viewHistoryBtn.addEventListener("click", () => activateView("history"));
}

if (historySubjectFilter) {
  historySubjectFilter.addEventListener("change", renderHistory);
}
if (historyStartInput) {
  historyStartInput.addEventListener("change", renderHistory);
}
if (historyEndInput) {
  historyEndInput.addEventListener("change", renderHistory);
}
if (historyClearBtn) {
  historyClearBtn.addEventListener("click", () => {
    if (historySubjectFilter) historySubjectFilter.value = "all";
    if (historyStartInput) historyStartInput.value = "";
    if (historyEndInput) historyEndInput.value = "";
    renderHistory();
  });
}

if (showRegisterBtn) {
  showRegisterBtn.addEventListener("click", () => setAuthMode("register"));
}

if (showLoginBtn) {
  showLoginBtn.addEventListener("click", () => setAuthMode("login"));
}

if (loginForm) {
  loginForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const email = document.getElementById("login-email").value.trim();
    const password = document.getElementById("login-password").value;
    showMessage(loginErrorEl, "");
    try {
      await fetchJSON("/api/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      await loadCurrentUser();
      if (isAuthenticated) {
        await loadAuthedData();
      }
    } catch (error) {
      showMessage(loginErrorEl, error.message || "Login failed");
    }
  });
}

if (registerForm) {
  registerForm.addEventListener("submit", async (event) => {
    event.preventDefault();
    const email = document.getElementById("register-email").value.trim();
    const password = document.getElementById("register-password").value;
    showMessage(registerErrorEl, "");
    try {
      await fetchJSON("/api/auth/register", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });
      await loadCurrentUser();
      if (isAuthenticated) {
        await loadAuthedData();
      }
    } catch (error) {
      showMessage(registerErrorEl, error.message || "Registration failed");
    }
  });
}

if (googleLoginBtn) {
  googleLoginBtn.addEventListener("click", async () => {
    try {
      const result = await fetchJSON("/api/auth/google/login", { headers: {} });
      if (result?.url) {
        window.location.href = result.url;
      }
    } catch (error) {
      alert(error.message || "Unable to start Google login");
    }
  });
}

if (logoutBtn) {
  logoutBtn.addEventListener("click", async () => {
    try {
      await fetchJSON("/api/auth/logout", { method: "POST", headers: {} });
    } finally {
      setAuthenticated(null);
    }
  });
}

async function initialize() {
  setDefaultTimes();
  setAuthMode("login");
  await loadCurrentUser();
  if (isAuthenticated) {
    await loadAuthedData();
  } else {
    activateView("auth", { skipSave: true, force: true });
  }
}

try {
  initLiveTrack();
} catch (error) {
  console.error("Failed to initialize live tracking UI", error);
}
initialize();
