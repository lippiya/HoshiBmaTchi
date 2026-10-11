<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from "vue";
import { useRouter } from "vue-router";

const router = useRouter();

const errors = ref({
  user: "",
  password: "",
  newPassword: "",
  confPassword: "",
});

const currentStep = ref(1);
const userEmail = ref("");

const form = ref({
  user: "",
  password: "",
  turnstile_token: "",
  otp: "",
  newPassword: "",
  confPassword: "",
});
const loading = ref(false);
const error = ref("");
const authToken = ref("");
const success = ref("");
const turnstileLoaded = ref(false);
const turnstileWidgetId = ref<string | null>(null);

const rateLimitSeconds = ref(0);
const rateLimitInterval = ref<number | null>(null);

onMounted(async () => {
  await nextTick();

  if (document.querySelector("script[src*=\"turnstile\"]")) {
    turnstileLoaded.value = true;
    renderTurnstile();
    return;
  }

  const script = document.createElement("script");
  script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js";
  script.async = true;
  script.defer = true;
  script.onload = () => {
    turnstileLoaded.value = true;
    renderTurnstile();
  };
  document.head.appendChild(script);
});

onUnmounted(() => {
  if (turnstileWidgetId.value && window.turnstile) {
    window.turnstile.reset(turnstileWidgetId.value);
    turnstileWidgetId.value = null;
  }
});

const validatePassword = (password: string): string[] => {
  const errors: string[] = [];
  if (!password) return errors;
  if (password.length < 8) errors.push("Password must be at least 8 characters");
  if (!/[A-Z]/.test(password)) errors.push("Password must contain at least one uppercase letter");
  if (!/[0-9]/.test(password)) errors.push("Password must contain at least one number");
  if (!/[!@#$%^&*(),.?":{}|<>]/.test(password))
    errors.push("Password must contain at least one symbol");
  return errors;
};

const validateField = (field: string) => {
  switch (field) {
    case "user":
      if (form.value.user.trim() === "") {
        errors.value.user = "User is required";
      } else {
        errors.value.user = "";
      }
      break;

    case "password":
      const pwdErrors = validatePassword(form.value.password);
      errors.value.password = pwdErrors[0] || "";
      break;
  }
};

const otpCountdown = (seconds: number) => {
  if (rateLimitInterval.value) {
    clearInterval(rateLimitInterval.value);
  }

  rateLimitSeconds.value = seconds;

  rateLimitInterval.value = setInterval(() => {
    rateLimitSeconds.value--;

    if (rateLimitSeconds.value <= 0) {
      if (rateLimitInterval.value) {
        clearInterval(rateLimitInterval.value);
        rateLimitInterval.value = null;
      }
    }
  }, 1000) as unknown as number;
};

const renderTurnstile = async () => {
  await nextTick();
  const container = document.querySelector("#turnstile-container");

  if (!container || !window.turnstile || turnstileWidgetId.value) {
    return;
  }

  if (container.innerHTML.trim() !== "") {
    return;
  }

  try {
    turnstileWidgetId.value = window.turnstile.render("#turnstile-container", {
      sitekey: import.meta.env.VITE_TURNSTILE_API_KEY,
      callback: (token: string) => {
        form.value.turnstile_token = token;
      },
      "error-callback": () => {
        error.value = "Turnstile verification failed. Please refresh the page.";
      },
    });
  } catch (err) {
    console.error("Turnstile render error:", err);
  }
};

const forgotPassword = () => {
  currentStep.value = 3;
};

const verifyOTP = async () => {
  error.value = "";
  success.value = "";

  if (!form.value.otp || form.value.otp.length !== 6) {
    error.value = "Please enter a valid 6-digit OTP code";
    return;
  }

  loading.value = true;

  try {
    const requestBody = {
      email: userEmail.value || form.value.user,
      code: form.value.otp,
    };

    const otpResponse = await fetch("https://localhost/api/v1/users/verify-otp", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestBody),
    });

    const data = await otpResponse.json();

    if (otpResponse.ok) {
      if (data.token) {
        authToken.value = data.token;
        localStorage.setItem("token", data.token);
      }

      success.value = "Email verified successfully!";
      setTimeout(() => {
        router.push("/");
      }, 1500);
    } else {
      error.value = data.error || "Invalid OTP code. Please try again.";
    }
  } catch (err) {
    error.value = "Network error. Please try again.";
    console.error("OTP verification error:", err);
  } finally {
    loading.value = false;
  }
};

const resendOTP = async () => {
  if (rateLimitSeconds.value > 0) {
    return;
  }
  error.value = "";
  success.value = "";
  loading.value = true;

  const emailToSend = userEmail.value || form.value.user;

  if (!emailToSend) {
    error.value = "Email address not found. Please try again.";
    loading.value = false;
    return;
  }

  try {
    const response = await fetch("https://localhost/api/v1/users/send-otp", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email: emailToSend }),
    });

    const data = await response.json();

    if (response.ok) {
      success.value = "OTP code has been resent to your email!";
      rateLimitSeconds.value = 60;
      otpCountdown(60);
    } else {
      const errorMsg = data.error || "";
      const match = errorMsg.match(/wait (\d+) seconds/);

      if (match && match[1]) {
        const seconds = parseInt(match[1]);
        otpCountdown(seconds);
        error.value = `Please wait ${seconds} seconds before requesting a new OTP`;
      } else {
        error.value = errorMsg || "Failed to resend OTP. Please try again.";
      }
    }
  } catch (err) {
    error.value = "Network error. Please try again.";
    console.error("Resend OTP error:", err);
  } finally {
    loading.value = false;
  }
};

const newPassword = async () => {
  error.value = "";
  success.value = "";
  validateField("user");
  validateField("newPassword");
  validateField("confPassword");
  if (Object.values(errors.value).some(err => err !== "")) {
    error.value = "Please fix the errors before submitting.";
    return;
  }
  try {
    const response = await fetch("https://localhost/api/v1/users/reset-password", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        user: form.value.user,
        new_password: form.value.newPassword,
        conf_password: form.value.confPassword,
      }),
    });
    const data = await response.json();
    if (!response.ok) {
      error.value = data.error || "Password reset failed";
      return;
    }
    success.value = "Password has been reset successfully!";
    setTimeout(() => {
      currentStep.value = 1;
      error.value = "";
      success.value = "";
      form.value.user = "";
      form.value.password = "";
      form.value.newPassword = "";
      form.value.confPassword = "";
    }, 1500);
  } catch (err) {
    error.value = "Network error. Please try again.";
    console.error("Reset password error:", err);
  }
};

const redirectRegister = () => {
  router.push("/register");
};

const submit = async () => {
  error.value = "";
  success.value = "";
  validateField("user");
  validateField("password");

  if (Object.values(errors.value).some(err => err !== "")) {
    error.value = "Please fix the errors before submitting.";
    return;
  }

  if (!form.value.turnstile_token) {
    error.value = "Please complete the verification.";
    return;
  }

  loading.value = true;

  try {
    const response = await fetch("/api/v1/users/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        email: form.value.user,
        password: form.value.password,
        turnstile_token: form.value.turnstile_token,
      }),
    });

    const contentType = response.headers.get("content-type");
    if (!contentType || !contentType.includes("application/json")) {
      error.value = "Service is unavailable. Please try again later.";
      loading.value = false;
      return;
    }

    const data = await response.json();

    if (!response.ok) {
      error.value = data.error || "Login failed";
      loading.value = false;
      return;
    }

    if (data.user && data.user.two_fa_enabled) {
      userEmail.value = data.user.email;
      currentStep.value = 2;
      success.value = "Please verify your 2FA code";
      resendOTP();
    } else {
      if (data.token) {
        localStorage.setItem("token", data.token);
      }
      success.value = "Login successful!";
      setTimeout(() => {
        router.push("/");
      }, 1000);
    }
  } catch (err) {
    error.value = "Network error. Please try again.";
    console.error("Login error:", err);
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="login-wrapper">
    <div class="login-container">
      <div class="login-card">
        <img
          src="../components/icons/logo-dark.png"
          alt="Logo"
          width="100%"
          height="100"
        />

        <div>
          <form
            v-if="currentStep === 1"
            id="login-form"
            @submit.prevent="submit"
          >
            <div class="form-group">
              <input
                id="username"
                v-model="form.user"
                type="text"
                required
                placeholder="Username"
                :disabled="loading"
              />
              <br />
              <span
                v-if="errors.user"
                class="error-text"
              >{{ errors.user }}</span>
            </div>

            <div class="form-group">
              <input
                id="password"
                v-model="form.password"
                type="password"
                required
                placeholder="Password"
                :disabled="loading"
                @blur="validateField('password')"
              />
              <br />
              <span
                v-if="errors.password"
                class="error-text"
              >{{ errors.password }}</span>
            </div>

            <div class="form-group">
              <div id="turnstile-container" />
            </div>

            <div
              v-if="error"
              class="alert alert-error"
              style="color: red"
            >
              {{ error }}
            </div>
            <div
              v-if="success"
              class="alert alert-success"
              style="color: greenyellow"
            >
              {{ success }}
            </div>
            <button
              type="submit"
              class="btn-primary"
              :disabled="loading"
              style="font-size: 18px; font-weight: 600"
            >
              Login
            </button>
            <div class="divider">
              <span>OR</span>
            </div>

            <h4
              style="color: purple; font-weight: 500"
              @click="forgotPassword"
            >
              Forgot your password?
            </h4>
            <p class="footer-text">
              Don't have an account?
              <router-link
                to="/register"
                style="color: purple"
              >
                Register here
              </router-link>
            </p>
          </form>

          <div
            v-else-if="currentStep === 2"
            class="otp-section"
          >
            <h2>Two-Factor Authentication</h2>
            <form @submit.prevent="verifyOTP">
              <div class="form-group">
                <input
                  v-model="form.otp"
                  type="text"
                  maxlength="6"
                  pattern="[0-9]{6}"
                  required
                  placeholder="Enter 6 digit OTP"
                  :disabled="loading"
                  class="otp-input"
                />
              </div>

              <div
                v-if="error"
                class="error-message"
              >
                {{ error }}
              </div>
              <div
                v-if="success"
                class="success-message"
              >
                {{ success }}
              </div>

              <button
                type="submit"
                class="btn-primary"
                :disabled="loading || form.otp.length !== 6"
              >
                {{ loading ? 'Verifying...' : 'Verify OTP' }}
              </button>
            </form>

            <div class="otp-actions">
              <button
                class="btn-link"
                :disabled="loading || rateLimitSeconds > 0"
                @click="resendOTP"
              >
                {{ rateLimitSeconds > 0 ? `Resend OTP (${rateLimitSeconds}s)` : 'Resend OTP' }}
              </button>
            </div>
          </div>
          <div v-if="currentStep === 3">
            <h2 style="font-weight: 500; margin-bottom: 20px">
              Reset Password
            </h2>
            <form @submit.prevent="newPassword">
              <div class="form-group">
                <input
                  id="username"
                  v-model="form.user"
                  type="text"
                  required
                  placeholder="Username or email"
                  :disabled="loading"
                />
                <br />
                <span
                  v-if="errors.user"
                  class="error-text"
                >{{ errors.user }}</span>
              </div>
              <div class="form-group">
                <input
                  v-model="form.newPassword"
                  type="password"
                  required
                  placeholder="New Password"
                  :disabled="loading"
                  @blur="validateField('newPassword')"
                />
                <br />
                <span
                  v-if="errors.newPassword"
                  class="error-text"
                >{{ errors.newPassword }}</span>
              </div>

              <div class="form-group">
                <input
                  v-model="form.confPassword"
                  type="password"
                  required
                  placeholder="Confirm New Password"
                  :disabled="loading"
                  @blur="validateField('confPassword')"
                />
                <br />
                <span
                  v-if="errors.confPassword"
                  class="error-text"
                >{{ errors.confPassword }}</span>
              </div>

              <div
                v-if="error"
                class="alert alert-error"
                style="color: red"
              >
                {{ error }}
              </div>
              <div
                v-if="success"
                class="alert alert-success"
                style="color: greenyellow"
              >
                {{ success }}
              </div>
              <button
                type="submit"
                class="btn-primary"
                :disabled="loading"
                style="font-size: 18px; font-weight: 600"
              >
                Reset Password
              </button>
              <button
                class="btn-link"
                :disabled="loading"
                style="margin-top: 10px"
                @click="currentStep = 1"
              >
                Back to Login
              </button>

              <button
                class="btn-link"
                :disabled="loading"
                style="margin-top: 10px"
                @click="redirectRegister"
              >
                Back to Register
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
* {
  color: white;
  background-color: black;
}

.login-wrapper {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  width: 100%;
  background-color: black;
}

.login-container {
  border: 1px solid gainsboro;
  padding: 20px;
  border-radius: 20px;
  width: 30vw;
  display: flex;
  flex-direction: column;
  justify-content: center;
  text-align: center;
  align-items: center;
}

.form-group input {
  border: white 1px solid;
  border-radius: 10px;
  width: 75%;
  padding: 10px;
  background-color: black;
  color: white;
  height: 30px;
  margin-bottom: 10px;
}

.form-group {
  margin-bottom: 15px;
}

.error-text {
  color: red;
  font-size: small;
}

button {
  width: 80%;
  height: 30px;
  border-radius: 5px;
  border: none;
  cursor: pointer;
  background-color: blueviolet;
}

.checkbox-group {
  display: flex;
  justify-content: center;
  align-items: center;
}

.checkbox-group label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.checkbox-group input[type='checkbox'] {
  width: auto;

  margin: 0;
  cursor: pointer;
}

.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 30px 0;
  gap: 10px;
}

.step {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  border: 2px solid #555;
  background-color: black;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: bold;
  transition: all 0.3s ease;
}

.step.active {
  border-color: blueviolet;
  background-color: blueviolet;
}

.step.completed {
  border-color: green;
  background-color: green;
}

.step-line {
  width: 80px;
  height: 2px;
  background-color: #555;
  transition: all 0.3s ease;
}

.step-line.active {
  background-color: blueviolet;
}

.otp-section {
  text-align: center;
}

.otp-info {
  margin: 15px 0;
  color: #ccc;
}

.otp-input {
  font-size: 20px;

  text-align: center;
}

.otp-actions {
  margin-top: 20px;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.btn-link {
  background: none;
  color: blueviolet;
  text-decoration: underline;
  border: none;
  cursor: pointer;
  width: auto;
}

.btn-link:hover {
  color: #a855f7;
}

.error-message {
  font-weight: 500;
  color: red;
  margin: 10px 0;
}

.success-message {
  font-weight: 500;
  color: greenyellow;
  margin: 10px 0;
}
</style>
