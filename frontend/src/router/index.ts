import { createWebHistory, createRouter } from "vue-router";
import { RouteRecordRaw } from "vue-router";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    alias: "/search",
    name: "search",
    component: () => import("@/components/ListPacks.vue"),
  },
  {
    path: "/package/:id",
    name: "package-data",
    component: () => import("@/components/PackDetails.vue"),
  },
  {
    path: "/upload",
    name: "npmIngest",
    component: () => import("@/components/NpmPackIngest.vue"),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
