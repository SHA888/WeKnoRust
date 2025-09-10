import { createRouter, createWebHistory } from 'vue-router'
import { checkInitializationStatus } from '@/api/initialization'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: "/",
      redirect: "/platform",
    },
    {
      path: "/initialization",
      name: "initialization",
      component: () => import("../views/initialization/InitializationConfig.vue"),
      meta: { requiresInit: false } // Initialization page does not require init check
    },
    {
      path: "/knowledgeBase",
      name: "home",
      component: () => import("../views/knowledge/KnowledgeBase.vue"),
      meta: { requiresInit: true }
    },
    {
      path: "/platform",
      name: "Platform",
      redirect: "/platform/knowledgeBase",
      component: () => import("../views/platform/index.vue"),
      meta: { requiresInit: true },
      children: [
        {
          path: "knowledgeBase",
          name: "knowledgeBase",
          component: () => import("../views/knowledge/KnowledgeBase.vue"),
          meta: { requiresInit: true }
        },
        {
          path: "creatChat",
          name: "creatChat",
          component: () => import("../views/creatChat/creatChat.vue"),
          meta: { requiresInit: true }
        },
        {
          path: "chat/:chatid",
          name: "chat",
          component: () => import("../views/chat/index.vue"),
          meta: { requiresInit: true }
        },
        {
          path: "settings",
          name: "settings",
          component: () => import("../views/settings/Settings.vue"),
          meta: { requiresInit: true }
        },
      ],
    },
  ],
});

// Route guard: check system initialization status
router.beforeEach(async (to, from, next) => {
  // If visiting the initialization page, allow directly
  if (to.meta.requiresInit === false) {
    next();
    return;
  }

  try {
    // Check if the system has been initialized
    const { initialized } = await checkInitializationStatus();
    
    if (initialized) {
      // System initialized: record to localStorage and proceed
      localStorage.setItem('system_initialized', 'true');
      next();
    } else {
      // System not initialized: redirect to initialization page
      console.log('System not initialized, redirecting to /initialization');
      next('/initialization');
    }
  } catch (error) {
    console.error('Failed to check initialization status:', error);
    // If the check fails, assume initialization is required
    next('/initialization');
  }
});

export default router
