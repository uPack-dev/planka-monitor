const MOTION_DISABLED_CLASS = 'motion-disabled';
const MOTION_DISABLED_STORAGE_KEY = 'annonce-motion-disabled';

let didInitializeMotionPreference = false;

export function useMotionPreference() {
  const motionDisabled = useState('motion-disabled', () => false);

  if (import.meta.client && !didInitializeMotionPreference) {
    didInitializeMotionPreference = true;
    motionDisabled.value = readMotionDisabled();
    applyMotionDisabled(motionDisabled.value);
  }

  function setMotionDisabled(value) {
    const nextValue = Boolean(value);
    motionDisabled.value = nextValue;
    writeMotionDisabled(nextValue);
    applyMotionDisabled(nextValue);
  }

  function toggleMotionDisabled() {
    setMotionDisabled(!motionDisabled.value);
  }

  return {
    motionDisabled,
    setMotionDisabled,
    toggleMotionDisabled,
  };
}

function readMotionDisabled() {
  try {
    return window.localStorage.getItem(MOTION_DISABLED_STORAGE_KEY) === '1';
  } catch {
    return false;
  }
}

function writeMotionDisabled(value) {
  try {
    window.localStorage.setItem(MOTION_DISABLED_STORAGE_KEY, value ? '1' : '0');
  } catch {
    // Local storage can be unavailable in restricted WebViews.
  }
}

function applyMotionDisabled(value) {
  document.documentElement.classList.toggle(MOTION_DISABLED_CLASS, value);
  document.body?.classList.toggle(MOTION_DISABLED_CLASS, value);
}
