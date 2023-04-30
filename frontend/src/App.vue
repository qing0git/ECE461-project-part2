<template>
  <div id="app">
    <nav class="navbar navbar-expand-sm navbar-dark bg-dark">
      <div class="container-fluid">
        <button class="navbar-toggler" type="button" data-bs-toggle="collapse" data-bs-target="#topnavbar">
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="collapse navbar-collapse" id="topnavbar">
          <ul class="navbar-nav me-auto">
            <li class="nav-item">
              <router-link to="/search" class="nav-link">Search</router-link>
            </li>
            <li class="nav-item">
              <router-link to="/upload" class="nav-link">Upload</router-link>
            </li>
          </ul>
          <div class="d-flex justify-content-center"><a class="navbar-brand">PM461</a></div>
          <button @click="resetPackRegistry" class="btn d-flex btn-danger navbar-btn" aria-label="Reset Data">Reset</button>
        </div>
      </div>
    </nav>
    <div class="mt-3">
      <router-view/>
    </div>
  </div>
</template>

<style>
@import '@/assets/style.css';
</style>

<script lang="ts">
import PackService from "@/services/PackService";

export default {
  methods: {
    resetPackRegistry() {
      if (confirm("Are you sure?")) {
        PackService.resetRegistry()
        .then((response) => {
            alert(response.data.message);
          })
        .catch((error) => {
          alert(error.response.data.message);
        })
        .finally(() => {
          window.location.reload();
        });
      }
    },
  },
};
</script>