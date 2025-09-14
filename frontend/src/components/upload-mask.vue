<script setup lang="ts">
</script>
<template>
    <div class="mask">
        <img class="upload-mask-img" src="@/assets/img/upload-mask.svg" alt="">
        <span class="drag-txt">Drag and drop files here</span>
        <span class="drag-type-txt">PDF, DOC files up to 30MB</span>
        <span class="drag-type-txt">Text, Markdown files up to 30MB</span>
    </div>
</template>
<style scoped lang="less">
.mask{
    display: flex;
    flex-flow: column;
    justify-content: center;
    align-items: center;
}
.drag-txt {
    color: #07c05f;
    font-family: "PingFang SC";
    font-size: 24px;
    font-weight: 600;
    line-height: 26px;
    display: inline-block;
    margin: 12px 0 16px 0;
}

.drag-type-txt {
    width: 217px;
    color: #00000066;
    text-align: center;
    font-family: "PingFang SC";
    font-size: 12px;
    font-weight: 400;
}
.upload-img{
    width: 162px;
    height: 162px;
}
</style>
