<script>
export default {
    props: ['username', 'userId', 'photoUrl'],
    emits: ['update-name', 'update-photo', 'logout'],
    data() {
        return {
            isEditingName: false,
            editNameValue: this.username || "",
        }
    },
    watch: {
        username(val) {
            this.editNameValue = val;
        }
    },
    methods: {
        saveName() {
            if (this.editNameValue !== this.username) {
                this.$emit('update-name', this.editNameValue);
            }
            this.isEditingName = false;
        },
        openPhotoUpload() {
             this.$refs.fileInput.click();
        },
        handleFileChange(event) {
            const file = event.target.files[0];
            if (file) {
                this.$emit('update-photo', file);
            }
        }
    }
}
</script>

<template>
    <div class="d-flex flex-column h-100 bg-dark-list border-end border-secondary">
        <input type="file" ref="fileInput" class="d-none" accept="image/*" @change="handleFileChange">
        <!-- Header -->
        <div class="p-3 bg-dark-header">
            <h5 class="m-0 text-white">Profile</h5>
        </div>

        <div class="flex-grow-1 p-3 overflow-auto">
            <!-- Photo -->
            <div class="d-flex justify-content-center mb-4 mt-3">
                <div class="position-relative" style="width: 150px; height: 150px; cursor: pointer;" @click="openPhotoUpload" title="Change Photo">
                    <div class="w-100 h-100 rounded-circle bg-secondary d-flex align-items-center justify-content-center overflow-hidden">
                        <!-- Placeholder or actual image if we had one -->
                        <img v-if="photoUrl" :src="photoUrl" class="w-100 h-100" style="object-fit: cover;">
                        <span v-else class="text-white fs-1">{{ username ? username.charAt(0).toUpperCase() : '?' }}</span>
                    </div>
                     <div class="position-absolute top-0 start-0 w-100 h-100 rounded-circle d-flex align-items-center justify-content-center bg-dark bg-opacity-50 text-white opacity-0 hover-opacity-100 transition">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-camera"><path d="M23 19a2 2 0 0 1-2 2H3a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h4l2-3h6l2 3h4a2 2 0 0 1 2 2z"></path><circle cx="12" cy="13" r="4"></circle></svg>
                    </div>
                </div>
            </div>

            <!-- Name -->
            <div class="mb-4">
                <label class="text-success small mb-2">Your Name</label>
                <div v-if="!isEditingName" class="d-flex justify-content-between align-items-center">
                    <span class="text-white">{{ username }}</span>
                    <button class="btn btn-link text-secondary p-0" @click="isEditingName = true">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-edit-2"><path d="M17 3a2.828 2.828 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5L17 3z"></path></svg>
                    </button>
                </div>
                <div v-else class="d-flex gap-2">
                    <input v-model="editNameValue" class="form-control form-control-sm bg-dark-input text-white border-bottom border-0 rounded-0 p-0" @keyup.enter="saveName">
                    <button class="btn btn-link text-secondary p-0" @click="saveName">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-check"><polyline points="20 6 9 17 4 12"></polyline></svg>
                    </button>
                </div>
                <small class="text-secondary d-block mt-2">This is not your username or pin. This name will be visible to your WASAText contacts.</small>
            </div>

             <!-- About (Static for now) -->
             <div class="mb-4">
                <label class="text-success small mb-2">About</label>
                <div class="text-white">Hey there! I am using WASAText.</div>
            </div>

            <hr class="border-secondary">

             <button class="btn btn-outline-danger w-100" @click="$emit('logout')">
                 Logout
             </button>
        </div>
    </div>
</template>

<style scoped>
.bg-dark-list { background-color: #111b21; }
.bg-dark-header { background-color: #202c33; }
.bg-dark-input { background-color: transparent; border-bottom: 2px solid #00a884 !important ; }
.bg-dark-input:focus { box-shadow: none; }

.text-secondary { color: #8696a0 !important; }
.text-success { color: #00a884 !important; }

.hover-opacity-100:hover { opacity: 1 !important; }
.transition { transition: opacity 0.2s; }
</style>
