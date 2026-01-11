<script>
export default {
    props: ['activeTab'],
    emits: ['update:activeTab'],
    methods: {
        setTab(tab) {
            this.$emit('update:activeTab', tab);
        }
    }
}
</script>

<template>
    <div class="sidebar-rail d-flex flex-column align-items-center py-3 bg-dark-rail border-end border-secondary">
        <div 
            class="nav-icon mb-4" 
            :class="{ active: activeTab === 'chat' }" 
            title="Chats"
            @click="setTab('chat')"
        >
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-message-square"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path></svg>
        </div>
        
        <div 
            class="nav-icon mb-4" 
            :class="{ active: activeTab === 'profile' }" 
            title="Profile"
            @click="setTab('profile')"
        >
            <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="feather feather-user"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
        </div>

    </div>
</template>

<style scoped>
.sidebar-rail {
    width: 64px;
    height: 100%;
}
.bg-dark-rail {
    background-color: #202c33;
}
.nav-icon {
    width: 40px;
    height: 40px;
    display: flex;
    justify-content: center;
    align-items: center;
    cursor: pointer;
    color: #8696a0;
    border-radius: 8px;
    transition: background-color 0.2s, color 0.2s;
}
.nav-icon:hover {
    background-color: #374045;
    color: #d1d7db;
}
.nav-icon.active {
    background-color: #374045;
    color: #00a884; /* WhatsApp Green */
}
</style>
